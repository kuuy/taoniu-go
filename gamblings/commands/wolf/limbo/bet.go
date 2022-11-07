package limbo

import (
	"errors"
	"github.com/gammazero/workerpool"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"

	repositories "taoniu.local/gamblings/repositories/wolf/limbo"
)

type BetHandler struct {
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
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					h.Repository.UseProxy = c.Bool("proxy")

					amount, _ := strconv.ParseFloat(c.Args().Get(0), 64)
					if amount < 0.00000001 {
						return errors.New("amount not valid")
					}

					multiplier, _ := strconv.ParseFloat(c.Args().Get(1), 64)
					if multiplier <= 1.01 {
						return errors.New("multiplier not valid")
					}

					if err := h.place(amount, multiplier); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *BetHandler) place(amount float64, multiplier float64) error {
	log.Println("wolf limbo bet place...")
	wp := workerpool.New(1)
	defer wp.StopWait()

	for {
		request := &repositories.BetRequest{
			Currency: "trx",
			Game:     "limbo",
		}
		request.Amount = strconv.FormatFloat(amount, 'f', -1, 64)
		request.Multiplier = strconv.FormatFloat(multiplier, 'f', -1, 64)

		hash, result, state, err := h.Repository.Place(request)
		if err != nil {
			log.Println("result verify error", err)
			os.Exit(1)
		}
		if state {
			log.Println("lucky", hash, result)
			os.Exit(1)
		}
	}

	return nil
}
