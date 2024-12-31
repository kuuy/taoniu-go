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

type TriggersHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  AnalysisRepository *analysisRepositories.TriggersRepository
}

func NewTriggersCommand() *cli.Command {
  var h TriggersHandler
  return &cli.Command{
    Name:  "triggers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TriggersHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.AnalysisRepository = &analysisRepositories.TriggersRepository{
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

func (h *TriggersHandler) Flush() (err error) {
  log.Println("binance futures tasks analysis tradings triggers flush...")
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    config.LOCKS_TASKS_ANALYSIS_TRADINGS_TRIGGERS_FLUSH,
  )
  if !mutex.Lock(50 * time.Second) {
    return
  }
  h.AnalysisRepository.Flush(1)
  h.AnalysisRepository.Flush(2)
  return
}
