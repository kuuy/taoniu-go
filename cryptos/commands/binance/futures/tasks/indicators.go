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
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type IndicatorsHandler struct {
  Db                   *gorm.DB
  Rdb                  *redis.Client
  Ctx                  context.Context
  Nats                 *nats.Conn
  IndicatorsRepository *repositories.IndicatorsRepository
  SymbolsRepository    *repositories.SymbolsRepository
  ScalpingRepository   *repositories.ScalpingRepository
}

func NewIndicatorsCommand() *cli.Command {
  var h IndicatorsHandler
  return &cli.Command{
    Name:  "indicators",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = IndicatorsHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.IndicatorsRepository = &repositories.IndicatorsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      baseRepository := indicatorsRepositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.IndicatorsRepository.Atr = &indicatorsRepositories.AtrRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.Pivot = &indicatorsRepositories.PivotRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.Kdj = &indicatorsRepositories.KdjRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.Rsi = &indicatorsRepositories.RsiRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.StochRsi = &indicatorsRepositories.StochRsiRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.Zlema = &indicatorsRepositories.ZlemaRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.HaZlema = &indicatorsRepositories.HaZlemaRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.BBands = &indicatorsRepositories.BBandsRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.AndeanOscillator = &indicatorsRepositories.AndeanOscillatorRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.IchimokuCloud = &indicatorsRepositories.IchimokuCloudRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.SuperTrend = &indicatorsRepositories.SuperTrendRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.Smc = &indicatorsRepositories.SmcRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.VolumeMoving = &indicatorsRepositories.VolumeMovingRepository{BaseRepository: baseRepository}
      h.IndicatorsRepository.VolumeProfile = &indicatorsRepositories.VolumeProfileRepository{BaseRepository: baseRepository}
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
    },
  }
}

func (h *IndicatorsHandler) Flush(interval string, current int) error {
  log.Printf("flushing futures indicators [%s] (batch %d)...", interval, current)
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

      mutex := common.NewMutex(h.Rdb, h.Ctx, fmt.Sprintf(config.LOCKS_TASKS_INDICATORS_FLUSH, interval, symbol))
      if !mutex.Lock(5 * time.Second) {
        return
      }
      defer mutex.Unlock()

      h.IndicatorsRepository.Flush(symbol, interval)
    }(symbol)
  }
  wg.Wait()

  return nil
}
