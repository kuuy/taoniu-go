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
		binance.Spot().Account().Flush()
		binance.Spot().Margin().Isolated().Account().Flush()
		binance.Spot().Margin().Isolated().Orders().Open()
		binance.Spot().Margin().Isolated().Symbols().Flush()
	})
	c.AddFunc("@every 5m", func() {
		binance.Spot().Klines().FlushDaily(2)
		binance.Spot().Indicators().Daily().Pivot()
		binance.Spot().Indicators().Daily().Atr(14, 100)
		binance.Spot().Indicators().Daily().Zlema(14, 100)
		binance.Spot().Indicators().Daily().HaZlema(14, 100)
		binance.Spot().Indicators().Daily().Kdj(9, 3, 100)
		binance.Spot().Indicators().Daily().BBands(14, 100)
		binance.Spot().Strategies().Daily().Atr()
		binance.Spot().Strategies().Daily().Zlema()
		binance.Spot().Strategies().Daily().HaZlema()
		binance.Spot().Strategies().Daily().Kdj()
		binance.Spot().Strategies().Daily().BBands()
	})
	c.AddFunc("0 30 * * * *", func() {
		binance.Spot().Klines().FlushDaily(50)
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
