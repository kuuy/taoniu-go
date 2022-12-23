package spot

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	Repository        *repositories.TickersRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func NewTickersCommand() *cli.Command {
	var h TickersHandler
	return &cli.Command{
		Name:  "tickers",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = TickersHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.TickersRepository{
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.SymbolsRepository = &repositories.SymbolsRepository{
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
		},
	}
}

func (h *TickersHandler) Flush() error {
	log.Println("Tickers flush...")
	symbols := h.SymbolsRepository.Scan()
	log.Println(symbols)
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols) {
			j = len(symbols)
		}
		h.Repository.Flush(symbols[i:j])
	}

	return nil
}
