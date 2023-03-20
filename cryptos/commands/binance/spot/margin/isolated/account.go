package isolated

import (
	"context"
	"gorm.io/gorm"
	"log"
	"taoniu.local/cryptos/tasks"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type AccountHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
	var h AccountHandler
	return &cli.Command{
		Name:  "account",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = AccountHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.AccountRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "transfer",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Transfer(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "loan",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Loan(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "repay",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Repay(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "collect",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Collect(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "liquidate",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Liquidate(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "test",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Test(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *AccountHandler) Flush() error {
	log.Println("margin isolated account flush processing...")
	return h.Repository.Flush()
}

func (h *AccountHandler) Transfer() error {
	log.Println("margin isolated account transfer...")
	asset := "AAVE"
	symbol := "AAVEBUSD"
	quantity := 0.01
	from := "ISOLATED_MARGIN"
	to := "SPOT"
	transferId, err := h.Repository.Transfer(asset, symbol, from, to, quantity)
	if err != nil {
		return err
	}
	log.Println("transferId", transferId)
	return nil
}

func (h *AccountHandler) Loan() error {
	asset := "BUSD"
	symbol := "ATOMBUSD"
	amount := 0.01
	transferId, err := h.Repository.Loan(asset, symbol, amount, true)
	if err != nil {
		return err
	}
	log.Println("transferId", transferId)
	return nil
}

func (h *AccountHandler) Repay() error {
	asset := "BUSD"
	symbol := "ATOMBUSD"
	amount := 0.01
	transferId, err := h.Repository.Repay(asset, symbol, amount, true)
	if err != nil {
		return err
	}
	log.Println("transferId", transferId)
	return nil
}

func (h *AccountHandler) Collect() error {
	log.Println("margin isolated account collect...")
	return h.Repository.Collect()
}

func (h *AccountHandler) Liquidate() error {
	log.Println("margin isolated account liquidate...")
	return h.Repository.Liquidate()
}

func (h *AccountHandler) Test() error {
	binance := tasks.BinanceTask{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}
	binance.Spot().Cron().Hourly()
	return nil
}
