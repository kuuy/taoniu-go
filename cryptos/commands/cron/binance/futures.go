package binance

import (
  "context"
  "log"
  "sync"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/nats-io/nats.go"
  "github.com/robfig/cron/v3"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/tasks"
)

type FuturesHandler struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Ctx   context.Context
  Asynq *asynq.Client
  Nats  *nats.Conn
}

func NewFuturesCommand() *cli.Command {
  var h FuturesHandler
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FuturesHandler{
        Db:    common.NewDB(2),
        Rdb:   common.NewRedis(2),
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

  ansqContext := &common.AnsqClientContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Conn: h.Asynq,
    Nats: h.Nats,
  }

  binance := tasks.NewBinanceTask(ansqContext)

  c := cron.New()
  c.AddFunc("@every 5s", func() {
    binance.Futures().Tickers().Flush()
    binance.Futures().FundingRate().Flush()
    binance.Futures().Scalping().Plans().Flush("1m")
    binance.Futures().Scalping().Plans().Flush("15m")
    binance.Futures().Scalping().Plans().Flush("4h")
    binance.Futures().Scalping().Plans().Flush("1d")
    binance.Futures().Tradings().Scalping().Place()
    //binance.Futures().Tradings().Gambling().Ant().Place()
    //binance.Futures().Tradings().Gambling().Scalping().Place()
  })
  c.AddFunc("@every 15s", func() {
    binance.Futures().Klines().Flush("1m")
    binance.Futures().Klines().Flush("15m")
    binance.Futures().Klines().Flush("4h")
    binance.Futures().Klines().Flush("1d")
    binance.Futures().Tradings().Scalping().Flush()
    //binance.Futures().Tradings().Gambling().Ant().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Futures().Orders().Open()
    binance.Futures().Orders().Flush()
  })
  c.AddFunc("@every 3m", func() {
    //binance.Futures().Depth().Flush(1000)
    binance.Futures().Orders().Sync(time.Now().Add(-15*time.Minute).UnixMilli(), 20)
  })
  c.AddFunc("@every 5m", func() {
    //binance.Futures().Analysis().Flush()
  })
  c.AddFunc("@every 15m", func() {
    //binance.Futures().Klines().Clean()
    //binance.Futures().Strategies().Clean()
    //binance.Futures().Plans().Clean()
  })
  c.AddFunc("@hourly", func() {
    //binance.Futures().Cron().Hourly()
  })
  c.AddFunc("0 8,20 * * * *", func() {
  })
  c.AddFunc("0 */336 * * * 1", func() {
  })
  c.AddFunc("30 23 * * * *", func() {
    binance.Server().Time()
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
