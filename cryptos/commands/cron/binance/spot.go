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
        Db:    common.NewDB(1),
        Rdb:   common.NewRedis(1),
        Asynq: common.NewAsynqClient("BINANCE_SPOT"),
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

  //tradingview := tasks.TradingviewTask{
  //  Db:    h.Db,
  //  Rdb:   h.Rdb,
  //  Ctx:   h.Ctx,
  //  Asynq: h.Asynq,
  //}

  ansqContext := &common.AnsqClientContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Conn: h.Asynq,
  }

  binance := tasks.NewBinanceTask(ansqContext)

  c := cron.New()
  c.AddFunc("@every 5s", func() {
    binance.Spot().Account().Flush()
    binance.Spot().Tickers().Flush()
    //binance.Spot().Tradings().Launchpad().Place()
    binance.Spot().Tradings().Triggers().Place()
    binance.Spot().Tradings().Scalping().Place()
    //binance.Spot().Margin().Cross().Tradings().Triggers().Place()
  })
  c.AddFunc("@every 15s", func() {
    //binance.Spot().Tradings().Launchpad().Flush()
    binance.Spot().Tradings().Triggers().Flush()
    binance.Spot().Tradings().Scalping().Flush()
    //binance.Spot().Margin().Cross().Tradings().Triggers().Flush()
    //tradingview.Analysis().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Spot().Orders().Open()
    binance.Spot().Orders().Flush()
    binance.Spot().Positions().Flush()
  })
  c.AddFunc("@every 1m", func() {
    binance.Spot().Depth().Flush(1000)
    binance.Spot().Orders().Sync(time.Now().Add(-15*time.Minute).UnixMicro(), 20)
  })
  c.AddFunc("@every 5m", func() {
  })
  c.AddFunc("@every 15m", func() {
  })
  c.AddFunc("@hourly", func() {
    binance.Spot().Cron().Hourly()
    //binance.Savings().Products().Flush()
  })
  c.AddFunc("15 1,11,19 * * *", func() {
    //binance.Spot().Margin().Isolated().Account().Collect()
  })
  c.AddFunc("45 1,17 * * *", func() {
    //binance.Spot().Tradings().Earn()
  })
  c.AddFunc("30 23 * * *", func() {
    binance.Server().Time()
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
