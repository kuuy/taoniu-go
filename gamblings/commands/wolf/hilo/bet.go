package hilo

import (
	"errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
	"strings"
	repositories "taoniu.local/gamblings/repositories/wolf/hilo"
)

type BetSerial struct {
	Rule string
	Size int
}

type BetHandler struct {
	Hash       string
	Amount     float64
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
					&cli.Float64Flag{
						Name:    "amount",
						Aliases: []string{"a"},
						Value:   0.00000001,
					},
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					h.Amount = c.Float64("amount")
					h.Repository.UseProxy = c.Bool("proxy")
					if c.Args().Get(0) == "" {
						return errors.New("rules is empty")
					}
					rules := strings.Split(c.Args().Get(0), ",")

					if c.Args().Get(1) == "" {
						return errors.New("sizes is empty")
					}
					sizes := strings.Split(c.Args().Get(1), ",")

					if len(rules) != len(sizes) {
						return errors.New("rules sizes not match")
					}

					gene := make([]*BetSerial, len(rules))
					for i, rule := range rules {
						gene[i] = &BetSerial{
							Rule: rule,
						}
						gene[i].Size, _ = strconv.Atoi(sizes[i])
					}

					if err := h.place(gene, 0); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *BetHandler) place(gene []*BetSerial, offset int) error {
	log.Println("wolf hilo bet place...")

	var hash string
	var betValue float64
	var status, subNonce int
	var err error

	var limit int
	for i, serial := range gene {
		if i > offset {
			continue
		}
		limit += serial.Size
	}

	for {
		hash, betValue, subNonce, err = h.Repository.Status()
		if err != nil {
			return err
		}

		rule := gene[offset].Rule

		log.Println("hits gene serial", rule, h.Amount, gene[offset].Size, subNonce, limit)

		for i := subNonce; i < limit; i++ {
			log.Println("play", hash, rule, betValue, subNonce)
			betValue, status, err = h.Repository.Play(h.Amount, rule, betValue, subNonce)
			if err != nil {
				return h.place(gene, offset)
			}

			if status == 0 {
				h.Repository.Start(h.Amount, betValue, subNonce)
				return h.place(gene, 0)
			}

			subNonce++
		}

		if offset < len(gene)-1 {
			return h.place(gene, offset+1)
		}

		h.Repository.Finish()
		h.Repository.Start(0.00000001, betValue, subNonce)

		log.Println("lucky", hash)
		os.Exit(1)
	}

	return nil
}
