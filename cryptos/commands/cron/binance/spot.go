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

type SpotHandler struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Ctx   context.Context
  Asynq *asynq.Client
  Nats  *nats.Conn
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
        Ctx:   context.Background(),
        Asynq: common.NewAsynqClient("BINANCE_SPOT"),
        Nats:  common.NewNats(),
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
    Nats: h.Nats,
  }

  binance := tasks.NewBinanceTask(ansqContext)

  c := cron.New()
  c.AddFunc("@every 5s", func() {
    binance.Spot().Tickers().Flush()
    binance.Spot().Tradings().Scalping().Place()
    //binance.Spot().Tradings().Gambling().Ant().Place()
    //binance.Spot().Tradings().Gambling().Scalping().Place()
  })
  c.AddFunc("@every 15s", func() {
    binance.Spot().Tradings().Scalping().Flush()
    binance.Spot().Tradings().Gambling().Ant().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Spot().Orders().Open()
    binance.Spot().Orders().Flush()
    binance.Spot().Positions().Flush()
  })
  c.AddFunc("@every 3m", func() {
    //binance.Spot().Depth().Flush(1000)
    binance.Spot().Orders().Sync(time.Now().Add(-15*time.Minute).UnixMilli(), 20)
  })
  c.AddFunc("@every 5m", func() {
    //binance.Spot().Analysis().Flush()
  })
  c.AddFunc("@every 15m", func() {
    //binance.Spot().Klines().Clean()
    //binance.Spot().Strategies().Clean()
    //binance.Spot().Plans().Clean()
  })
  c.AddFunc("@hourly", func() {
    //binance.Spot().Symbols().Flush()
    //binance.Savings().Products().Flush()
  })
  c.AddFunc("0 8,20 * * * *", func() {
  })
  c.AddFunc("15 1,11,19 * * *", func() {
    //binance.Spot().Margin().Isolated().Account().Collect()
  })
  c.AddFunc("45 1,17 * * *", func() {
    //binance.Spot().Tradings().Earn()
  })
  c.AddFunc("30 23 * * *", func() {
    binance.Server().Time()
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
