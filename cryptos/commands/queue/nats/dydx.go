package nats

import (
  "context"
  "log"
  "sync"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers"
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

func (h *DydxHandler) Run() error {
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
  workers.NewDydx(natsContext).Subscribe()

  <-h.wait(wg)

  return nil
}

func (h *DydxHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
