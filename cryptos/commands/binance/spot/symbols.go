package spot

import (
	"context"
	"gorm.io/gorm"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
	tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type SymbolsHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
	var h SymbolsHandler
	return &cli.Command{
		Name:  "symbols",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = SymbolsHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.SymbolsRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.TradingsRepository = &repositories.TradingsRepository{
				Db: h.Db,
			}
			h.Repository.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "currencies",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Currencies(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
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
				Name:  "scan",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Scan(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "count",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Count(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "slippage",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Slippage(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "adjust",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Adjust(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *SymbolsHandler) Currencies() error {
	log.Println("symbols currencies...")
	currencies := h.Repository.Currencies()
	log.Println("currencies", currencies)
	return nil
}

func (h *SymbolsHandler) Flush() error {
	log.Println("symbols flush...")
	return h.Repository.Flush()
}

func (h *SymbolsHandler) Scan() error {
	log.Println("symbols scan...")
	symbols := h.Repository.Scan()
	log.Println("symbols", symbols)
	return nil
}

func (h *SymbolsHandler) Count() error {
	log.Println("symbols count...")
	return h.Repository.Count()
}

func (h *SymbolsHandler) Slippage() error {
	log.Println("symbols depth...")
	for _, symbol := range h.Repository.Symbols() {
		h.Repository.Slippage(symbol)
	}
	return nil
}

func (h *SymbolsHandler) Adjust() error {
	log.Println("symbols adjust...")
	symbol := "AVAXBUSD"
	price := 11.81 * 1.02
	price, quantity, err := h.Repository.Adjust(symbol, price, 20)
	log.Println("price", price, quantity, err)
	return nil
}
