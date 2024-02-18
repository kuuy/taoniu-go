package binance

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  workers "taoniu.local/cryptos/queue/asynq/workers/binance"
)

type FuturesHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewFuturesCommand() *cli.Command {
  var h FuturesHandler
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FuturesHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
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

func (h *FuturesHandler) Run() error {
  log.Println("queue running...")

  mux := asynq.NewServeMux()
  worker := common.NewAsynqServer("BINANCE_FUTURES")

  ansqContext := &common.AnsqServerContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Mux:  mux,
    Nats: common.NewNats(),
  }

  workers.NewFutures(ansqContext).Register()
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
