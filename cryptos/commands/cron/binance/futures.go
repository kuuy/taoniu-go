package binance

import (
  "context"
  "log"
  "sync"
  "time"

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
        Asynq: common.NewAsynqClient("BINANCE_FUTURES"),
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
  c.AddFunc("@every 5s", func() {
    binance.Futures().Account().Flush()
    binance.Futures().Tickers().Flush()
    binance.Futures().Tradings().Triggers().Place()
    binance.Futures().Tradings().Scalping().Place()
  })
  c.AddFunc("@every 15s", func() {
    binance.Futures().Tradings().Triggers().Flush()
    binance.Futures().Tradings().Scalping().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Futures().Orders().Open()
    binance.Futures().Orders().Flush()
  })
  c.AddFunc("@every 3m", func() {
    binance.Futures().Depth().Flush(1000)
    binance.Futures().Orders().Sync(time.Now().Add(-72*time.Hour).UnixMilli(), 200)
  })
  c.AddFunc("@every 5m", func() {
    binance.Futures().Klines().FlushDelay("1m", 30)
    binance.Futures().Klines().FlushDelay("15m", 2)
    binance.Futures().Depth().FlushDelay(1000)
    binance.Futures().Analysis().Flush()
  })
  c.AddFunc("@every 15m", func() {
    binance.Futures().Klines().FlushDelay("4h", 1)
    binance.Futures().Klines().FlushDelay("1d", 1)
  })
  c.AddFunc("@hourly", func() {
    binance.Futures().Cron().Hourly()
  })
  c.AddFunc("0 20 * * * *", func() {
    binance.Futures().Scalping().Flush()
  })
  c.AddFunc("0 */336 * * * 1", func() {
    binance.Futures().Triggers().Flush()
  })
  c.AddFunc("30 23 * * * *", func() {
    binance.Server().Time()
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
