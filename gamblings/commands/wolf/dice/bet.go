package dice

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	pool "taoniu.local/gamblings/common"
	wolfRepositories "taoniu.local/gamblings/repositories/wolf"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type BetHandler struct {
	Rdb               *redis.Client
	Ctx               context.Context
	Mode              string
	IPart             string
	DPart             string
	Numbers           []float64
	Repository        *repositories.BetRepository
	HuntRepository    *repositories.HuntRepository
	AccountRepository *wolfRepositories.AccountRepository
}

func NewBetCommand() *cli.Command {
	var h BetHandler
	return &cli.Command{
		Name:  "bet",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = BetHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.BetRepository{
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.HuntRepository = &repositories.HuntRepository{
				Db: pool.NewDB(),
			}
			h.AccountRepository = &wolfRepositories.AccountRepository{
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "place",
				Usage: "",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "mode",
						Value: "",
					},
					&cli.StringFlag{
						Name:    "numbers",
						Aliases: []string{"n"},
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
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					h.Mode = c.String("mode")
					if c.String("numbers") != "" {
						numbers := strings.Split(c.String("numbers"), ",")
						h.Numbers = make([]float64, len(numbers))
						for i := 0; i < len(numbers); i++ {
							h.Numbers[i], _ = strconv.ParseFloat(numbers[i], 64)
						}
					}
					h.IPart = c.String("ipart")
					h.DPart = c.String("dpart")
					h.Repository.UseProxy = c.Bool("proxy")
					h.AccountRepository.UseProxy = c.Bool("proxy")
					if err := h.place(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "test",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.test(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "multiple",
				Usage: "",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					amount, _ := strconv.ParseFloat(c.Args().Get(0), 64)
					if amount < 0.00000001 {
						return errors.New("amount not valid")
					}
					amount = math.Ceil(amount/0.00000001) / math.Ceil(1/0.00000001)
					h.Repository.UseProxy = c.Bool("proxy")
					if err := h.multiple(amount); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *BetHandler) multiple(amount float64) error {
	log.Println("wolf dice multiple bet...")

	amount = math.Ceil(amount*0.001/0.00000001) / math.Ceil(1/0.00000001)

	var profitAmount float64 = 0
	var lossAmount float64 = 0
	var winAmount float64 = 0
	var count = 0
	var fails = 0

	rules := []string{"under", "over"}
	for {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })
		rule := rules[(rand.Intn(571-23)+23)%len(rules)]
		var betValue float64
		if rule == "under" {
			betValue = 67
		} else {
			betValue = 33
		}

		_, _, state, err := h.Repository.Place("trx", amount, rule, betValue)
		if err != nil {
			log.Println("bet error", err)
			continue
		}

		multiplier, _ := h.Repository.BetRule(rule, betValue)

		if state {
			winAmount += amount * (multiplier - 1)
		} else {
			lossAmount += amount
			fails++
		}

		count++
		profitAmount = winAmount - lossAmount

		if profitAmount > 0 {
			break
		}

		if fails >= 3 {
			amount = math.Ceil(-profitAmount/(3*0.99*(multiplier-1))/0.00000001) / math.Ceil(1/0.00000001)
			fails = 0
		}

		time.Sleep(1 * time.Second)
	}

	profitAmount = math.Ceil(profitAmount/0.00000001) / math.Ceil(1/0.00000001)
	winAmount = math.Ceil(winAmount/0.00000001) / math.Ceil(1/0.00000001)
	lossAmount = math.Ceil(lossAmount/0.00000001) / math.Ceil(1/0.00000001)

	log.Println(
		"profit",
		count,
		strconv.FormatFloat(amount, 'f', -1, 64),
		strconv.FormatFloat(profitAmount, 'f', -1, 64),
		strconv.FormatFloat(winAmount, 'f', -1, 64),
		strconv.FormatFloat(lossAmount, 'f', -1, 64),
	)

	return nil
}

func (h *BetHandler) test() error {
	var result float64
	result = 34.34
	h.Mode = "repeate"
	//h.Numbers = []float64{11.22, 3.33}
	//h.IPart = "13-23"
	if !h.verify(result) {
		log.Println("result verify false")
	} else {
		log.Println("result verify ok")
	}

	return nil
}

func (h *BetHandler) verify(result float64) bool {
	number := strconv.FormatFloat(result, 'f', -1, 64)
	parts := strings.Split(number, ".")
	if len(parts) == 1 {
		parts = append(parts, "0")
	}

	if len(h.Numbers) > 0 {
		for _, item := range h.Numbers {
			if result == item {
				return true
			}
		}
		return false
	}

	if h.IPart == "o" {
		number, _ := strconv.Atoi(parts[0])
		if number%2 != 1 {
			return false
		}
	} else if h.IPart == "e" {
		number, _ := strconv.Atoi(parts[0])
		if number%2 != 0 {
			return false
		}
	} else if h.IPart != "" {
		ranges := strings.Split(h.IPart, "-")
		if len(ranges) == 2 {
			number, _ := strconv.Atoi(parts[0])
			min, _ := strconv.Atoi(ranges[0])
			max, _ := strconv.Atoi(ranges[1])
			if number < min || number > max {
				return false
			}
		} else {
			numbers := strings.Split(h.IPart, ",")
			contains := false
			for _, number := range numbers {
				if parts[0] == number {
					contains = true
					break
				}
			}
			if !contains {
				return false
			}
		}
	}

	if h.DPart == "o" {
		number, _ := strconv.Atoi(parts[1])
		if number%2 != 1 {
			return false
		}
	} else if h.DPart == "e" {
		number, _ := strconv.Atoi(parts[1])
		if number%2 != 0 {
			return false
		}
	} else if h.DPart != "" {
		ranges := strings.Split(h.DPart, "-")
		if len(ranges) == 2 {
			number, _ := strconv.Atoi(parts[1])
			min, _ := strconv.Atoi(ranges[0])
			max, _ := strconv.Atoi(ranges[1])
			if number < min || number > max {
				return false
			}
		} else {
			numbers := strings.Split(h.DPart, ",")
			contains := false
			for _, number := range numbers {
				if parts[1] == number {
					contains = true
					break
				}
			}
			if !contains {
				return false
			}
		}
	}

	if h.Mode == "same" {
		if parts[0] == "0" {
			parts[0] = "00"
		}
		if parts[1] == "0" {
			parts[1] = "00"
		}

		if len(parts[0]) != 2 || len(parts[1]) != 2 {
			return false
		}

		if parts[0][0] != parts[0][1] || parts[1][0] != parts[1][1] {
			return false
		}
	}

	if h.Mode == "repeate" {
		if parts[0] != parts[1] {
			return false
		}
	}

	if h.Mode == "same-up" {
		if parts[0] == "0" {
			parts[0] = "00"
		}

		if len(parts[0]) != 2 || len(parts[1]) != 2 {
			return false
		}

		if parts[0][0] != parts[0][1] || parts[1][0] != parts[1][1] || parts[0][0] != parts[1][0]-1 {
			return false
		}
	}

	if h.Mode == "same-down" {
		if parts[1] == "0" {
			parts[1] = "00"
		}

		if len(parts[0]) != 2 || len(parts[1]) != 2 {
			return false
		}

		if parts[0][0] != parts[0][1] || parts[1][0] != parts[1][1] || parts[0][0] != parts[1][0]+1 {
			return false
		}
	}

	if h.Mode == "mirror" {
		if len(parts[0]) != 2 || len(parts[1]) != 2 {
			return false
		}

		if parts[0][0] == parts[0][1] || parts[1][0] == parts[1][1] {
			return false
		}

		if parts[0][0] == parts[1][1] && parts[0][1] == parts[1][0] {
			return true
		} else {
			return false
		}
	}

	if h.Mode == "asc" {
		if len(parts[0])+len(parts[1]) != 3 {
			return false
		}
		start := parts[0][0]
		for i := 1; i < len(parts[0]); i++ {
			if parts[0][i] != start+1 {
				return false
			}
			start += 1
		}
		for i := 0; i < len(parts[1]); i++ {
			if parts[1][i] != start+1 {
				return false
			}
			start += 1
		}
	}

	if h.Mode == "dec" {
		if len(parts[0])+len(parts[1]) != 3 {
			return false
		}
		start := parts[0][0]
		for i := 1; i < len(parts[0]); i++ {
			if parts[0][i] != start-1 {
				return false
			}
			start -= 1
		}
		for i := 0; i < len(parts[1]); i++ {
			if parts[1][i] != start-1 {
				return false
			}
			start -= 1
		}
	}

	if h.Mode == "neighbor" {
		if len(parts[0]) < 2 {
			parts[0] = "0" + parts[0]
		}
		if len(parts[1]) < 2 {
			parts[1] = parts[1] + "0"
		}

		start := parts[0][0]
		for i := 1; i < len(parts[0]); i++ {
			if parts[0][0] > parts[0][1] {
				if parts[0][i] != start-1 {
					return false
				}
				start -= 1
			} else {
				if parts[0][i] != start+1 {
					return false
				}
				start += 1
			}
		}

		for i := 0; i < len(parts[1]); i++ {
			if parts[0][0] > parts[0][1] {
				if parts[1][i] != start-1 {
					return false
				}
				start -= 1
			} else {
				if parts[1][i] != start+1 {
					return false
				}
				start += 1
			}
		}
	}

	return false
}

func (h *BetHandler) place() error {
	log.Println("wolf dice bet place...")

	for {
		hash, result, _, err := h.Repository.Place("trx", 0.000001, "under", 98)
		if err != nil {
			log.Println("result verify error", err)
			os.Exit(1)
		}
		h.HuntRepository.Handing(hash, result)

		if h.verify(result) {
			log.Println("lucky", hash, result)
			os.Exit(1)
		}
	}

	return nil
}
