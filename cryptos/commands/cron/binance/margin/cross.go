package margin

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

type CrossHandler struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Ctx   context.Context
  Asynq *asynq.Client
  Nats  *nats.Conn
}

func NewCrossCommand() *cli.Command {
  var h CrossHandler
  return &cli.Command{
    Name:  "cross",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CrossHandler{
        Db:    common.NewDB(1),
        Rdb:   common.NewRedis(1),
        Ctx:   context.Background(),
        Asynq: common.NewAsynqClient("BINANCE_MARGIN_CROSS"),
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

func (h *CrossHandler) run() error {
  log.Println("binance margin cross cron running...")

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
    binance.Margin().Cross().Tradings().Scalping().Place()
    binance.Margin().Cross().Tradings().Triggers().Place()
  })
  c.AddFunc("@every 15s", func() {
    binance.Margin().Cross().Tradings().Scalping().Flush()
    binance.Margin().Cross().Tradings().Triggers().Flush()
  })
  c.AddFunc("@every 30s", func() {
    binance.Margin().Cross().Orders().Open()
    binance.Margin().Cross().Orders().Flush()
    binance.Margin().Cross().Positions().Flush()
  })
  c.AddFunc("@every 1m", func() {
    binance.Margin().Cross().Orders().Sync(time.Now().Add(-15*time.Minute).UnixMicro(), 20)
  })
  c.AddFunc("@every 5m", func() {
    binance.Margin().Cross().Analysis().Flush()
  })
  c.AddFunc("@every 15m", func() {
  })
  c.AddFunc("@hourly", func() {
    //binance.Savings().Products().Flush()
  })
  c.AddFunc("0 8,20 * * * *", func() {
  })
  c.AddFunc("15 1,11,19 * * *", func() {
    //binance.Margin().Cross().Margin().Isolated().Account().Collect()
  })
  c.AddFunc("45 1,17 * * *", func() {
    //binance.Margin().Cross().Tradings().Earn()
  })
  c.AddFunc("30 23 * * *", func() {
    //binance.Server().Time()
  })
  c.Start()

  <-h.wait(wg)

  return nil
}

func (h *CrossHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
