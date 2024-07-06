package margin

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  workers "taoniu.local/cryptos/queue/asynq/workers/binance/margin"
)

type CrossHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewCrossCommand() *cli.Command {
  var h CrossHandler
  return &cli.Command{
    Name:  "cross",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CrossHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.Run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *CrossHandler) Run() error {
  log.Println("queue running...")

  mux := asynq.NewServeMux()
  worker := common.NewAsynqServer("BINANCE_MARGIN_CROSS")

  ansqContext := &common.AnsqServerContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Mux:  mux,
    Nats: common.NewNats(),
  }

  workers.NewCross(ansqContext).Register()
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
