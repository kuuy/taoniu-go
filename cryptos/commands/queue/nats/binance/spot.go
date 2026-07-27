package binance

import (
  "context"
  "log"
  "sync"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  workers "taoniu.local/cryptos/queue/nats/workers/binance"
)

type SpotHandler struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Nats *nats.Conn
}

func NewSpotCommand() *cli.Command {
  var h SpotHandler
  return &cli.Command{
    Name:  "spot",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SpotHandler{
        Db:   common.NewDB(1),
        Rdb:  common.NewRedis(1),
        Ctx:  context.Background(),
        Nats: common.NewNats(1),
      }
      return nil
    },
    After: func(c *cli.Context) error {
      sqlDB, _ := h.Db.DB()
      sqlDB.Close()
      h.Rdb.Close()
      h.Nats.Close()
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

func (h *SpotHandler) Run() error {
  log.Println("nats running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  natsContext := &common.NatsContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Conn: h.Nats,
  }
  workers.NewSpot(natsContext).Subscribe()

  <-h.wait(wg)

  return nil
}

func (h *SpotHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
