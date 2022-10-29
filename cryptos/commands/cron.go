package commands

import (
	"context"
	"log"
	"sync"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	"taoniu.local/cryptos/tasks"
)

type CronHandler struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func NewCronCommand() *cli.Command {
	var h CronHandler
	return &cli.Command{
		Name:  "cron",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = CronHandler{
				Db:  pool.NewDB(),
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
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

func (h *CronHandler) run() error {
	log.Println("cron running...")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	binance := tasks.BinanceTask{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}

	c := cron.New()
	c.AddFunc("@every 30s", func() {
		binance.Symbols().Flush()
		binance.Spot().Flush()
		binance.Spot().Margin().Isolated().Tradings().Grids()
		binance.Spot().Margin().Isolated().Tradings().UpdateGrids()
		binance.Spot().Tradings().Scalping()
		binance.Spot().Tradings().UpdateScalping()
	})
	c.AddFunc("@every 3m", func() {
		binance.Spot().Margin().Orders().Fix()
		binance.Spot().Orders().Fix()
	})
	c.AddFunc("@every 5m", func() {
		binance.Spot().Indicators().Daily().Flush()
		binance.Spot().Strategies().Daily().Flush()
		binance.Spot().Analysis().Daily().Flush()
	})
	c.AddFunc("@hourly", func() {
		binance.Symbols().Count()
		binance.Spot().Margin().Sync()
		binance.Spot().Sync()
	})
	c.AddFunc("0 30 * * * *", func() {
		binance.Spot().Plans().Daily().Fix()
		binance.Spot().Clean()
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
