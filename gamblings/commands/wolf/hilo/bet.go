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

	amount := 0.00000001

	for {
		hash, betValue, subNonce, err = h.Repository.Status()
		if err != nil {
			return err
		}

		for i := subNonce; i < limit; i++ {
			log.Println("play", hash, betValue, subNonce)
			betValue, status, err = h.Repository.Play(amount, rule, betValue, subNonce)
			if err != nil {
				return h.place(rule, limit)
			}

			if status == 0 {
				h.Repository.Start(amount, betValue, subNonce)
				return h.place(rule, limit)
			}

			subNonce++
		}

		h.Repository.Finish()
		h.Repository.Start(amount, betValue, subNonce)

		log.Println("lucky", hash)
		os.Exit(1)
	}

	return nil
}
