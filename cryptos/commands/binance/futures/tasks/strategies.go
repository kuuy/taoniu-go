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
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
  strategiesRepositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type StrategiesHandler struct {
  Db                   *gorm.DB
  Rdb                  *redis.Client
  Ctx                  context.Context
  StrategiesRepository *repositories.StrategiesRepository
  ScalpingRepository   *repositories.ScalpingRepository
}

func NewStrategiesCommand() *cli.Command {
  var h StrategiesHandler
  return &cli.Command{
    Name:  "strategies",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = StrategiesHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.StrategiesRepository = &repositories.StrategiesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      baseIndicatorsRepository := indicatorsRepositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      baseStrategiesRepository := strategiesRepositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.StrategiesRepository.Atr = &strategiesRepositories.AtrRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.AtrRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.Kdj = &strategiesRepositories.KdjRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.KdjRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.Rsi = &strategiesRepositories.RsiRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.RsiRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.StochRsi = &strategiesRepositories.StochRsiRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.StochRsiRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.Zlema = &strategiesRepositories.ZlemaRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.ZlemaRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.HaZlema = &strategiesRepositories.HaZlemaRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.HaZlemaRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.BBands = &strategiesRepositories.BBandsRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.BBandsRepository{BaseRepository: baseIndicatorsRepository},
      }
      h.StrategiesRepository.IchimokuCloud = &strategiesRepositories.IchimokuCloudRepository{
        BaseRepository: baseStrategiesRepository,
        Repository:     &indicatorsRepositories.IchimokuCloudRepository{BaseRepository: baseIndicatorsRepository},
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

func (h *StrategiesHandler) Flush(interval string, current int) error {
  log.Printf("flushing futures strategies [%s] (batch %d)...", interval, current)
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

  var wg sync.WaitGroup
  semaphore := make(chan struct{}, 10)

  for _, symbol := range symbols[startPos:endPos] {
    wg.Add(1)
    go func(s string) {
      defer wg.Done()
      semaphore <- struct{}{}
      defer func() { <-semaphore }()

      mutex := common.NewMutex(h.Rdb, h.Ctx, fmt.Sprintf(config.LOCKS_TASKS_STRATEGIES_FLUSH, interval, symbol))
      if !mutex.Lock(5 * time.Second) {
        return
      }
      defer mutex.Unlock()

      h.StrategiesRepository.Flush(symbol, interval)
    }(symbol)
  }
  wg.Wait()

  return nil
}

func (h *StrategiesHandler) Clean() error {
  log.Println("binance futures tasks strategies clean...")
  symbols := h.ScalpingRepository.Scan(2)
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_STRATEGIES_CLEAN, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.StrategiesRepository.Clean(symbol)
  }
  return nil
}
