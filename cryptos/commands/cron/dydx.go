package cron

import (
  "context"
  "log"
  "sync"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/robfig/cron/v3"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/tasks"
)

type DydxHandler struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Asynq *asynq.Client
  Ctx   context.Context
}

func NewDydxCommand() *cli.Command {
  var h DydxHandler
  return &cli.Command{
    Name:  "dydx",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DydxHandler{
        Db:    common.NewDB(1),
        Rdb:   common.NewRedis(1),
        Asynq: common.NewAsynqClient("DYDX"),
        Ctx:   context.Background(),
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
  log.Println("cron running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  dydx := tasks.DydxTask{
    Db:    h.Db,
    Rdb:   h.Rdb,
    Ctx:   h.Ctx,
    Asynq: h.Asynq,
  }

  c := cron.New()
  c.AddFunc("@every 5s", func() {
    dydx.Account().Flush()
    dydx.Tickers().Flush()
    dydx.Orderbook().Flush()
    dydx.Tradings().Triggers().Place()
    dydx.Tradings().Scalping().Place()
  })
  c.AddFunc("@every 15s", func() {
    dydx.Tradings().Triggers().Flush()
    dydx.Tradings().Scalping().Flush()
  })
  c.AddFunc("@every 30s", func() {
    dydx.Orders().Open()
  })
  c.AddFunc("@every 5m", func() {
    dydx.Analysis().Flush()
    //dydx.Orderbook().FlushDelay()
    //dydx.Klines().FlushDelay("1m", 30)
    //dydx.Klines().FlushDelay("15m", 2)
  })
  c.AddFunc("@every 15m", func() {
    //dydx.Klines().FlushDelay("4h", 1)
    //dydx.Klines().FlushDelay("1d", 1)
  })
  c.AddFunc("@hourly", func() {
    dydx.Cron().Hourly()
  })
  c.AddFunc("0 20 * * * *", func() {
    dydx.Scalping().Flush()
  })
  c.AddFunc("0 */336 * * * 1", func() {
    dydx.Triggers().Flush()
  })
  c.AddFunc("30 23 * * *", func() {
    dydx.Server().Time()
    dydx.Clean()
  })
  c.Start()

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
