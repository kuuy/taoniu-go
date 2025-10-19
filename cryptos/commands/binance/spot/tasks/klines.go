package tasks

import (
  "context"
  "errors"
  "fmt"
  "log"
  "slices"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
        Db:   common.NewDB(1),
        Rdb:  common.NewRedis(1),
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
            log.Fatal("interval not valid")
            return nil
          }
          current, _ := strconv.Atoi(c.Args().Get(1))
          if current < 1 {
            log.Fatal("current is less than 1")
            return nil
          }
          if err := h.Flush(interval, current); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
            log.Fatal("interval not valid")
            return nil
          }
          current, _ := strconv.Atoi(c.Args().Get(1))
          if current < 1 {
            log.Fatal("current is less than 1")
            return nil
          }
          if err := h.Fix(interval, current); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "clean",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Clean(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *KlinesHandler) Flush(interval string, current int) (err error) {
  log.Println("binance spot tasks klines flush...", interval, current)
  symbols := h.ScalpingRepository.Scan()

  pageSize := common.GetEnvInt("BINANCE_SPOT_SYMBOLS_SIZE")
  startPos := (current - 1) * pageSize
  if startPos >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := startPos + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  timestamp := h.KlinesRepository.Timestamp(interval)

  for _, symbol := range symbols[startPos:endPos] {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_FLUSH, interval, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }

    duration := time.Second * 0
    if interval == "1m" {
      duration = time.Second * (30 + 60)
    } else if interval == "15m" {
      duration = time.Second * (30 + 900)
    } else if interval == "4h" {
      duration = time.Second * (30 + 14400)
    } else if interval == "1d" {
      duration = time.Second * (30 + 86400)
    }

    redisKey := fmt.Sprintf(config.REDIS_KEY_KLINES, interval, symbol, timestamp)
    data, _ := h.Rdb.HMGet(
      h.Ctx,
      redisKey,
      "open",
      "close",
      "high",
      "low",
      "volume",
      "quota",
      "lasttime",
    ).Result()
    if data[0] != nil &&
      data[1] != nil &&
      data[2] != nil &&
      data[3] != nil &&
      data[4] != nil &&
      data[5] != nil &&
      data[6] != nil {
      open, _ := strconv.ParseFloat(data[0].(string), 64)
      close, _ := strconv.ParseFloat(data[1].(string), 64)
      high, _ := strconv.ParseFloat(data[2].(string), 64)
      low, _ := strconv.ParseFloat(data[3].(string), 64)
      volume, _ := strconv.ParseFloat(data[4].(string), 64)
      quota, _ := strconv.ParseFloat(data[5].(string), 64)
      lasttime, _ := strconv.ParseInt(data[6].(string), 10, 64)

      entity, err := h.KlinesRepository.Get(symbol, interval, timestamp)
      if errors.Is(err, gorm.ErrRecordNotFound) {
        h.KlinesRepository.Create(
          symbol,
          interval,
          open,
          close,
          high,
          low,
          volume,
          quota,
          timestamp,
        )
      } else if lasttime > entity.UpdatedAt.UnixMilli() {
        h.KlinesRepository.Updates(entity, map[string]interface{}{
          "open":   open,
          "close":  close,
          "high":   high,
          "low":    low,
          "volume": volume,
          "quota":  quota,
        })
      } else {
        diff := time.Now().UnixMilli() - entity.UpdatedAt.UnixMilli()
        if diff > 30000 {
          log.Println("flush klines", symbol, interval, diff)
          h.KlinesRepository.Flush(symbol, interval, 0, 1)
        }
      }
      ttl, _ := h.Rdb.TTL(h.Ctx, redisKey).Result()
      if -1 == ttl.Nanoseconds() {
        h.Rdb.Expire(h.Ctx, redisKey, duration)
      }
    }
    //h.KlinesRepository.Flush(symbol, "1m", 0, 5)
    //h.KlinesRepository.Flush(symbol, "15m", 0, 1)
    //h.KlinesRepository.Flush(symbol, "4h", 0, 1)
    //h.KlinesRepository.Flush(symbol, "1d", 0, 1)
  }
  return nil
}

func (h *KlinesHandler) Fix(interval string, current int) (err error) {
  log.Println("binance spot tasks klines fix...", interval, current)
  symbols := h.ScalpingRepository.Scan()

  pageSize := common.GetEnvInt("BINANCE_SPOT_SYMBOLS_SIZE")
  startPos := (current - 1) * pageSize
  if startPos >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := startPos + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else if interval == "1d" {
    limit = 100
  }

  for _, symbol := range symbols[startPos:endPos] {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_FIX, interval, symbol),
    )
    if !mutex.Lock(30 * time.Second) {
      continue
    }
    h.KlinesRepository.Fix(symbol, interval, limit)
  }
  return nil
}

func (h *KlinesHandler) Clean() error {
  log.Println("binance spot tasks klines clean...")
  symbols := h.ScalpingRepository.Scan()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_CLEAN, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.KlinesRepository.Clean(symbol)
  }
  return nil
}
