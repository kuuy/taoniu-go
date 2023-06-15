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

type SpotHandler struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Asynq *asynq.Client
  Ctx   context.Context
}

func NewSpotCommand() *cli.Command {
  var h SpotHandler
  return &cli.Command{
    Name:  "spot",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SpotHandler{
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

func (h *SpotHandler) run() error {
  log.Println("cron running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  tradingview := tasks.TradingviewTask{
    Db:    h.Db,
    Rdb:   h.Rdb,
    Ctx:   h.Ctx,
    Asynq: h.Asynq,
  }

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
    binance.Spot().Tickers().Flush()
    binance.Spot().Tradings().Scalping().Place()
    binance.Spot().Tradings().Triggers().Place()
    binance.Spot().Tradings().Fishers().Place()
    binance.Spot().Margin().Cross().Tradings().Triggers().Place()
    binance.Spot().Margin().Isolated().Tradings().Fishers().Place()
  })
  c.AddFunc("@every 15s", func() {
    binance.Spot().Tradings().Fishers().Flush()
    binance.Spot().Tradings().Scalping().Flush()
    binance.Spot().Margin().Cross().Tradings().Triggers().Flush()
    binance.Spot().Margin().Isolated().Tradings().Fishers().Flush()
    tradingview.Analysis().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Spot().Flush()
    binance.Spot().Depth().Flush()
  })
  c.AddFunc("@every 1m", func() {
    binance.Spot().Klines().Flush("1m", 5)
    binance.Spot().Klines().Flush("1d", 1)
  })
  c.AddFunc("@every 3m", func() {
    binance.Spot().Tickers().Fix()
    binance.Spot().Klines().Fix("1m", 30, 270)
    binance.Spot().Klines().Fix("1d", 2, 2700)
    binance.Spot().Margin().Orders().Fix()
    binance.Spot().Orders().Fix()
  })
  c.AddFunc("@every 5m", func() {
    binance.Spot().Indicators().Daily().Flush()
    binance.Spot().Strategies().Daily().Flush()
    binance.Spot().Analysis().Flush()
    binance.Spot().Depth().FlushDelay()
    binance.Spot().Tickers().FlushDelay()
    binance.Spot().Klines().FlushDelay("1m", 30)
    tradingview.Analysis().FlushDelay()
    //binance.Futures().Indicators().Daily().Place()
    //binance.Futures().Strategies().Daily().Place()
  })
  c.AddFunc("@every 15m", func() {
    binance.Spot().Klines().FlushDelay("1d", 1)
  })
  c.AddFunc("@hourly", func() {
    binance.Spot().Cron().Hourly()
    //binance.Futures().Cron().Hourly()
    binance.Savings().Products().Flush()
  })
  c.AddFunc("15 1,11,19 * * *", func() {
    binance.Spot().Margin().Isolated().Account().Collect()
  })
  c.AddFunc("45 1,17 * * *", func() {
    binance.Spot().Tradings().Collect()
    binance.Spot().Margin().Isolated().Tradings().Fishers().Grids().Collect()
  })
  c.AddFunc("30 23 * * *", func() {
    binance.Spot().Clean()
  })
  c.Start()

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
