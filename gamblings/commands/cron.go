package commands

import (
	"context"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"

	"taoniu.local/gamblings/common"
	"taoniu.local/gamblings/tasks"
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
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
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

	wolf := tasks.WolfTask{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}

	c := cron.New()
	c.AddFunc("@every 5s", func() {
		wolf.Account().Flush()
	})
	c.AddFunc("@every 30s", func() {
		var err error
		mask := 7
		for {
			if mask == 0 {
				break
			}

			rand.Seed(time.Now().UnixNano())
			i := (rand.Intn(571-23) + 23) % 3
			if i == 0 && mask&1 == 1 {
				err = wolf.Dice().Multiple().Apply("usdt")
				if err != nil {
					mask ^= 1
					continue
				}
			} else if i == 1 && mask&2 == 2 {
				err = wolf.Dice().Plans().Apply("usdt")
				if err != nil {
					mask ^= 2
					continue
				}
			} else if i == 2 && mask&4 == 4 {
				err = wolf.Dice().Hells().Apply("usdt")
				if err != nil {
					mask ^= 4
					continue
				}
			}
		}
	})
	c.AddFunc("30 2 * * *", func() {
		wolf.Dice().Multiple().Rescue()
		wolf.Dice().Plans().Rescue()
		wolf.Dice().Hells().Rescue()
	})
	c.AddFunc("45 7,15 * * *", func() {
		wolf.Dice().Bet().Start()
	})
	c.AddFunc("15 1,11,19 * * *", func() {
		wolf.Dice().Bet().Stop()
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
