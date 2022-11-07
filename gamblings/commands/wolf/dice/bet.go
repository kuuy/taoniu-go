package dice

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type BetHandler struct {
	Mode       string
	IPart      string
	DPart      string
	Numbers    []float64
	Repository *repositories.BetRepository
}

func NewBetCommand() *cli.Command {
	var h BetHandler
	return &cli.Command{
		Name:  "bet",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = BetHandler{}
			h.Repository = &repositories.BetRepository{}
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
		},
	}
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

	return true
}

func (h *BetHandler) place() error {
	log.Println("wolf dice bet place...")

	for {
		request := &repositories.BetRequest{
			Currency:   "trx",
			Game:       "dice",
			Multiplier: "1.0102",
			Amount:     "0.000001",
			Rule:       "under",
			BetValue:   98,
		}
		hash, result, _, err := h.Repository.Place(request)
		if err != nil {
			log.Println("result verify error", err)
			os.Exit(1)
		}

		if h.verify(result) {
			log.Println("lucky", hash, result)
			os.Exit(1)
		}
	}

	return nil
}
