package tradings

import (
  "context"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  analysisRepositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
)

type ScalpingHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  AnalysisRepository *analysisRepositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.AnalysisRepository = &analysisRepositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Flush() (err error) {
  log.Println("binance futures tasks analysis tradings scalping flush...")
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    config.LOCKS_TASKS_ANALYSIS_TRADINGS_SCALPING_FLUSH,
  )
  if !mutex.Lock(50 * time.Second) {
    return
  }
  h.AnalysisRepository.Flush(1)
  h.AnalysisRepository.Flush(2)
  return
}
