package commands

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"sync"
	"taoniu.local/cryptos/tasks"

	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
)

type CronHandler struct {
	db  *gorm.DB
	rdb *redis.Client
	ctx context.Context
}

func NewCronCommand() *cli.Command {
	h := CronHandler{
		db:  pool.NewDB(),
		rdb: pool.NewRedis(),
		ctx: context.Background(),
	}
	return &cli.Command{
		Name:  "cron",
		Usage: "",
		Action: func(c *cli.Context) error {
			if err := h.run(); err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return nil
		},
	}
}

func (h *CronHandler) run() error {
	log.Println("cron running...")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	binance := tasks.BinanceTask{
		Db:  h.db,
		Rdb: h.rdb,
		Ctx: h.ctx,
	}

	c := cron.New()
	c.AddFunc("@every 30s", func() {
		binance.Spot().Flush()
	})
	c.AddFunc("@every 3m", func() {
		binance.Spot().Margin().Orders().Fix()
	})
	c.AddFunc("@every 5m", func() {
		binance.Spot().Klines().Daily().Flush(1)
		binance.Spot().Indicators().Daily().Flush()
		binance.Spot().Strategies().Daily().Flush()
	})
	c.AddFunc("@hourly", func() {
		binance.Spot().Margin().Sync()
		binance.Spot().Analysis().Daily().Flush()
	})
	c.AddFunc("0 30 * * * *", func() {
		binance.Spot().Symbols().Flush()
		binance.Spot().Klines().Daily().Flush(2)
	})
	c.Start()

	<-h.wait(wg)

	return nil
}

func (h *CronHandler) wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}
