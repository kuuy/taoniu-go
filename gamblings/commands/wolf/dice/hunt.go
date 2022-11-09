package dice

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	common "taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type HuntCondition struct {
	Numbers    string
	Ipart      string
	Dpart      string
	Side       string
	IsMirror   bool
	IsRepeate  bool
	IsNeighbor bool
}

type HuntHandler struct {
	Rdb           *redis.Client
	Ctx           context.Context
	HuntCondition *HuntCondition
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
			h.HuntCondition = &HuntCondition{}
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
			{
				Name:  "stop",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.stop(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "monitor",
				Usage: "",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "numbers",
						Aliases: []string{"n"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "side",
						Aliases: []string{"s"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "ipart",
						Aliases: []string{"i"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "dpart",
						Aliases: []string{"d"},
						Value:   "",
					},
					&cli.BoolFlag{
						Name:    "mirror",
						Aliases: []string{"b1"},
						Value:   false,
					},
					&cli.BoolFlag{
						Name:    "repeate",
						Aliases: []string{"b2"},
						Value:   false,
					},
					&cli.BoolFlag{
						Name:    "neighbor",
						Aliases: []string{"b3"},
						Value:   false,
					},
				},
				Action: func(c *cli.Context) error {
					h.HuntCondition.Numbers = c.String("numbers")
					h.HuntCondition.Side = c.String("side")
					h.HuntCondition.Ipart = c.String("ipart")
					h.HuntCondition.Dpart = c.String("dpart")
					h.HuntCondition.IsMirror = c.Bool("mirror")
					h.HuntCondition.IsRepeate = c.Bool("repeate")
					h.HuntCondition.IsNeighbor = c.Bool("neighbor")
					if err := h.monitor(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *HuntHandler) place() error {
	log.Println("wolf dice hunt place...")

	wp := workerpool.New(5)
	defer wp.StopWait()

	amount := 0.00000001
	rules := []string{"under", "over"}
	for {
		timestamp := time.Now().Unix()
		score, _ := h.Rdb.ZScore(
			h.Ctx,
			"wolf:hunts",
			"dice",
		).Result()
		if int64(score) < timestamp-1800 {
			log.Println("hunt not start")
			break
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })
		rule := rules[(rand.Intn(571-23)+23)%len(rules)]
		var betValue float64
		if rule == "under" {
			betValue = 67
		} else {
			betValue = 33
		}
		log.Println("rule", rule, betValue)
		hash, result, _, err := h.BetRepository.Place(amount, rule, betValue)
		if err != nil {
			log.Println("bet error", err)
			continue
		}
		wp.Submit(func() {
			h.Repository.Handing(hash, result)
		})
	}

	return nil
}

func (h *HuntHandler) start() error {
	log.Println("wolf dice hunt starting...")
	return h.Repository.Start()
}

func (h *HuntHandler) stop() error {
	log.Println("wolf dice hunt stopping...")
	return h.Repository.Stop()
}

func (h *HuntHandler) monitor() error {
	log.Println("wolf dice hunt monitor...")
	conditions := make(map[string]interface{})

	if h.HuntCondition.Numbers != "" {
		values := strings.Split(h.HuntCondition.Numbers, ",")
		numbers := make([]float64, len(values))
		for i := 0; i < len(values); i++ {
			numbers[i], _ = strconv.ParseFloat(values[i], 64)
		}
		conditions["numbers"] = numbers
	}

	if h.HuntCondition.Side != "" {
		side, _ := strconv.Atoi(h.HuntCondition.Side)
		conditions["side"] = side
	}

	if h.HuntCondition.Ipart != "" {
		var numbers []int
		ranges := strings.Split(h.HuntCondition.Ipart, "-")
		if len(ranges) == 2 {
			min, _ := strconv.Atoi(ranges[0])
			max, _ := strconv.Atoi(ranges[1])
			for i := min; i < max; i++ {
				numbers = append(numbers, i)
			}
		} else {
			values := strings.Split(h.HuntCondition.Ipart, ",")
			for i := 0; i < len(values); i++ {
				value, _ := strconv.Atoi(values[i])
				numbers = append(numbers, value)
			}
		}
		conditions["ipart"] = numbers
	}

	if h.HuntCondition.Dpart != "" {
		var numbers []int
		ranges := strings.Split(h.HuntCondition.Dpart, "-")
		if len(ranges) == 2 {
			min, _ := strconv.Atoi(ranges[0])
			max, _ := strconv.Atoi(ranges[1])
			for i := min; i < max; i++ {
				numbers = append(numbers, i)
			}
		} else {
			values := strings.Split(h.HuntCondition.Dpart, ",")
			for i := 0; i < len(values); i++ {
				value, _ := strconv.Atoi(values[i])
				numbers = append(numbers, value)
			}
		}
		conditions["dpart"] = numbers
	}

	if h.HuntCondition.IsMirror {
		conditions["is_mirror"] = true
	}

	if h.HuntCondition.IsRepeate {
		conditions["is_repeate"] = true
	}

	if h.HuntCondition.IsNeighbor {
		conditions["is_neighbor"] = true
	}

	score, _ := h.Rdb.ZScore(
		h.Ctx,
		"wolf:hunts",
		"dice",
	).Result()
	if score == 0 {
		h.start()
	}
	conditions["opentime"] = time.Unix(int64(score), 0)

	for {
		hunts := h.Repository.Gets(conditions)
		for _, hunt := range hunts {
			log.Println("lucky", hunt.Number, hunt.Hash)
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}
