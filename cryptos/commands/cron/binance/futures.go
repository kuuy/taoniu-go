package binance

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

type FuturesHandler struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Asynq *asynq.Client
  Ctx   context.Context
}

func NewFuturesCommand() *cli.Command {
  var h FuturesHandler
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FuturesHandler{
        Db:    common.NewDB(),
        Rdb:   common.NewRedis(),
        Asynq: common.NewAsynqClient(),
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

func (h *FuturesHandler) run() error {
  log.Println("cron running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  binance := tasks.BinanceTask{
    Db:    h.Db,
    Rdb:   h.Rdb,
    Ctx:   h.Ctx,
    Asynq: h.Asynq,
  }

  c := cron.New()
  c.AddFunc("@every 1s", func() {
    binance.Server().Time()
  })
  c.AddFunc("@every 5s", func() {
    binance.Futures().Tickers().Flush()
    binance.Futures().Tradings().Triggers().Place()
  })
  c.AddFunc("@every 15s", func() {
    binance.Futures().Tradings().Triggers().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Futures().Orders().Open()
    binance.Futures().Flush()
  })
  c.AddFunc("@every 1m", func() {
    binance.Futures().Klines().Flush("1m", 5)
    binance.Futures().Klines().Flush("1d", 1)
  })
  c.AddFunc("@every 3m", func() {
    binance.Futures().Tickers().Fix()
    binance.Futures().Klines().Fix("1m", 30, 270)
    binance.Futures().Klines().Fix("1d", 2, 2700)
    binance.Futures().Orders().Fix()
  })
  c.AddFunc("@every 5m", func() {
    binance.Futures().Tickers().FlushDelay()
    binance.Futures().Klines().FlushDelay("1m", 30)
    binance.Futures().Indicators().Daily().Flush()
    binance.Futures().Strategies().Daily().Flush()
    binance.Futures().Indicators().Minutely().Flush()
    binance.Futures().Strategies().Minutely().Flush()
    binance.Futures().Plans().Minutely().Flush()
    //binance.Futures().Indicators().Daily().Place()
    //binance.Futures().Strategies().Daily().Place()
  })
  c.AddFunc("@every 15m", func() {
    binance.Futures().Klines().FlushDelay("1d", 1)
  })
  c.AddFunc("@hourly", func() {
    binance.Futures().Cron().Hourly()
  })
  c.AddFunc("30 23 * * *", func() {
    binance.Futures().Clean()
  })
  c.Start()

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
