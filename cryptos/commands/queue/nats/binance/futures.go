package binance

import (
  "context"
  "gorm.io/gorm"
  "log"
  "sync"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  workers "taoniu.local/cryptos/queue/nats/workers/binance"
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
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
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
  log.Println("nats running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  nc := common.NewNats()
  defer nc.Close()

  natsContext := &common.NatsContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Conn: nc,
  }
  workers.NewFutures(natsContext).Subscribe()

  <-h.wait(wg)

  return nil
}

func (h *FuturesHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
