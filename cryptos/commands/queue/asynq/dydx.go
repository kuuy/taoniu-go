package asynq

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers"
)

type DydxHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewDydxCommand() *cli.Command {
  var h DydxHandler
  return &cli.Command{
    Name:  "dydx",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DydxHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *DydxHandler) run() error {
  log.Println("queue running...")

  mux := asynq.NewServeMux()
  worker := common.NewAsynqServer("DYDX")

  ansqContext := &common.AnsqServerContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Mux:  mux,
    Nats: common.NewNats(),
  }

  workers.NewDydx(ansqContext).Register()
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
