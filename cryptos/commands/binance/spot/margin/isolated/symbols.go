package isolated

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type SymbolsHandler struct {
	Repository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
	h := SymbolsHandler{
		Repository: &repositories.SymbolsRepository{
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "symbols",
		Usage: "",
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *SymbolsHandler) flush() error {
	log.Println("symbols flush processing...")
	return h.Repository.Flush()
}
