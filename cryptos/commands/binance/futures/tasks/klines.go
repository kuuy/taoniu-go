package tasks

import (
  "context"
  "errors"
  "fmt"
  "log"
  "slices"
  "strconv"
  "sync"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type KlinesHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Nats               *nats.Conn
  KlinesRepository   *repositories.KlinesRepository
  SymbolsRepository  *repositories.SymbolsRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:   common.NewDB(2),
        Rdb:  common.NewRedis(2),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.KlinesRepository = &repositories.KlinesRepository{
        Db:   h.Db,
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Nats: h.Nats,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
            return errors.New("invalid interval")
          }
          current, _ := strconv.Atoi(c.Args().Get(1))
          if current < 1 {
            return errors.New("current index must be >= 1")
          }
          return h.Flush(interval, current)
        },
      },
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
            return errors.New("invalid interval")
          }
          current, _ := strconv.Atoi(c.Args().Get(1))
          if current < 1 {
            return errors.New("current index must be >= 1")
          }
          return h.Fix(interval, current)
        },
      },
      {
        Name:  "clean",
        Usage: "",
        Action: func(c *cli.Context) error {
          return h.Clean()
        },
      },
    },
  }
}

func (h *KlinesHandler) Flush(interval string, current int) error {
  log.Printf("flushing futures klines [%s] (batch %d)...", interval, current)
  symbols := h.ScalpingRepository.Scan(2)

  pageSize := common.GetEnvInt("BINANCE_FUTURES_SYMBOLS_SIZE")
  if pageSize <= 0 {
    pageSize = 100
  }
  startPos := (current - 1) * pageSize
  if startPos >= len(symbols) {
    return errors.New("symbols out of range")
  }
  endPos := startPos + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  timestamp := h.KlinesRepository.Timestamp(interval)

  var wg sync.WaitGroup
  semaphore := make(chan struct{}, 10)

  for _, symbol := range symbols[startPos:endPos] {
    wg.Add(1)
    go func(s string) {
      defer wg.Done()
      semaphore <- struct{}{}
      defer func() { <-semaphore }()

      mutex := common.NewMutex(h.Rdb, h.Ctx, fmt.Sprintf(config.LOCKS_TASKS_KLINES_FLUSH, interval, symbol))
      if !mutex.Lock(5 * time.Second) {
        return
      }
      defer mutex.Unlock()

      redisKey := fmt.Sprintf(config.REDIS_KEY_KLINES, interval, symbol, timestamp)
      fields := []string{"open", "close", "high", "low", "volume", "quota", "lasttime"}
      data, err := h.Rdb.HMGet(h.Ctx, redisKey, fields...).Result()
      if err != nil || len(data) != len(fields) {
        return
      }

      for _, v := range data {
        if v == nil {
          return
        }
      }

      open, _ := strconv.ParseFloat(data[0].(string), 64)
      closePrice, _ := strconv.ParseFloat(data[1].(string), 64)
      high, _ := strconv.ParseFloat(data[2].(string), 64)
      low, _ := strconv.ParseFloat(data[3].(string), 64)
      volume, _ := strconv.ParseFloat(data[4].(string), 64)
      quota, _ := strconv.ParseFloat(data[5].(string), 64)
      lasttime, _ := strconv.ParseInt(data[6].(string), 10, 64)

      entity, err := h.KlinesRepository.Get(symbol, interval, timestamp)
      if errors.Is(err, gorm.ErrRecordNotFound) {
        h.KlinesRepository.Create(symbol, interval, open, closePrice, high, low, volume, quota, timestamp)
      } else if lasttime > entity.UpdatedAt.UnixMilli() {
        h.KlinesRepository.Updates(entity, map[string]interface{}{
          "open":   open,
          "close":  closePrice,
          "high":   high,
          "low":    low,
          "volume": volume,
          "quota":  quota,
        })
      } else {
        if time.Now().UnixMilli()-entity.UpdatedAt.UnixMilli() > 30000 {
          h.KlinesRepository.Flush(symbol, interval, 0, 1)
        }
      }
    }(symbol)
  }
  wg.Wait()

  return nil
}

func (h *KlinesHandler) Fix(interval string, current int) error {
  log.Printf("fixing futures klines [%s] (batch %d)...", interval, current)
  symbols := h.ScalpingRepository.Scan(2)

  pageSize := common.GetEnvInt("BINANCE_FUTURES_SYMBOLS_SIZE")
  if pageSize <= 0 {
    pageSize = 50
  }
  startPos := (current - 1) * pageSize
  if startPos >= len(symbols) {
    return errors.New("symbols out of range")
  }
  endPos := startPos + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  var limit int
  switch interval {
  case "1m":
    limit = 1440
  case "15m":
    limit = 672
  case "4h":
    limit = 126
  case "1d":
    limit = 100
  default:
    limit = 100
  }

  for _, symbol := range symbols[startPos:endPos] {
    mutex := common.NewMutex(h.Rdb, h.Ctx, fmt.Sprintf(config.LOCKS_TASKS_KLINES_FIX, interval, symbol))
    if !mutex.Lock(30 * time.Second) {
      continue
    }
    h.KlinesRepository.Fix(symbol, interval, limit)
  }
  return nil
}

func (h *KlinesHandler) Clean() error {
  log.Println("cleaning old futures klines...")
  symbols := h.ScalpingRepository.Scan(2)
  for _, symbol := range symbols {
    mutex := common.NewMutex(h.Rdb, h.Ctx, fmt.Sprintf(config.LOCKS_TASKS_KLINES_CLEAN, symbol))
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.KlinesRepository.Clean(symbol)
    mutex.Unlock()
  }
  return nil
}
