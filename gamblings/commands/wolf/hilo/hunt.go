package hilo

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"

	common "taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf/hilo"
)

type HuntHandler struct {
	Rdb           *redis.Client
	Ctx           context.Context
	Repository    *repositories.HuntRepository
	BetRepository *repositories.BetRepository
}

func NewHuntCommand() *cli.Command {
	var h HuntHandler
	return &cli.Command{
		Name:  "hunt",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = HuntHandler{
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.HuntRepository{
				Db:  common.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.BetRepository = &repositories.BetRepository{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "place",
				Usage: "",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					h.BetRepository.UseProxy = c.Bool("proxy")
					if err := h.place(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.start(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *HuntHandler) place() error {
	log.Println("wolf hilo hunt place...")

	wp := workerpool.New(5)
	defer wp.StopWait()

	//for {
	//	timestamp := time.Now().Unix()
	//	score, _ := h.Rdb.ZScore(
	//		h.Ctx,
	//		"wolf:hunts",
	//		"hilo",
	//	).Result()
	//	if int64(score) < timestamp-1800 {
	//		log.Println("hunt not start")
	//		break
	//	}
	//
	//	request := &repositories.BetRequest{
	//		Currency:   "trx",
	//		Game:       "hilo",
	//		Multiplier: "1.0102",
	//		Amount:     "0.000001",
	//		Rule:       "under",
	//		BetValue:   98,
	//	}
	//	hash, result, _, err := h.BetRepository.Place(request)
	//	if err != nil {
	//		log.Println(" bet error", err)
	//		continue
	//	}
	//	wp.Submit(func() {
	//		h.Repository.Handing(hash, result)
	//	})
	//}

	return nil
}

func (h *HuntHandler) start() error {
	log.Println("wolf hilo hunt starting...")
	return h.Repository.Start()
}
