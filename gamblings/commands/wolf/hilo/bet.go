package hilo

import (
	"errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
	repositories "taoniu.local/gamblings/repositories/wolf/hilo"
)

type BetHandler struct {
	Hash       string
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
					rule := c.Args().Get(0)
					if rule == "" {
						return errors.New("rule is empty")
					}
					limit, _ := strconv.Atoi(c.Args().Get(1))
					if limit < 1 {
						return errors.New("limit not valid")
					}
					if err := h.place(rule, limit); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *BetHandler) place(rule string, limit int) error {
	log.Println("wolf hilo bet place...")

	var hash string
	var betValue float64
	var status, subNonce int
	var err error

	request := &repositories.BetRequest{
		Currency: "usdt",
		Amount:   "0.00000001",
		Rule:     rule,
	}
	if rule == "red" {
		request.Multiplier = "1.98"
		request.WinChance = 50
	} else if rule == "black" {
		request.Multiplier = "1.98"
		request.WinChance = 50
	} else if rule == "number" {
		request.Multiplier = "1.43"
		request.WinChance = 69.23
	} else if rule == "letter" {
		request.Multiplier = "3.2174"
		request.WinChance = 30.77
	} else {
		return errors.New("rule not supported")
	}

	for {
		if hash == "" {
			hash, betValue, subNonce, err = h.Repository.Status(request)
			if err != nil {
				return err
			}
		}

		request.BetValue = betValue
		request.SubNonce = subNonce
		betValue, status, subNonce, err = h.Repository.Place(request, limit)
		if err != nil {
			continue
		}
		if status == 0 {
			request.BetValue = betValue
			request.SubNonce = subNonce
			hash, betValue, err = h.Repository.Start(request)
			if err != nil {
				return err
			}
			continue
		}

		log.Println("lucky", hash)
		os.Exit(1)
	}

	return nil
}
