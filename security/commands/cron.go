package commands

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"log"
	"sync"

	"taoniu.local/security/common"
	"taoniu.local/security/tasks"
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
				Db: common.NewDB(),
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

	gfw := tasks.GfwTask{
		Db: h.Db,
	}

	c := cron.New()
	c.AddFunc("@every 30s", func() {
		gfw.Dns().Flush()
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
