package spot

import (
	"context"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"log"
	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
	Db         *gorm.DB
	Repository *repositories.TickersRepository
}

func NewTickersCommand() *cli.Command {
	h := TickersHandler{
		Db: pool.NewDB(),
		Repository: &repositories.TickersRepository{
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "tickers",
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

func (h *TickersHandler) flush() error {
	log.Println("Tickers flush...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for i := 0; i < len(symbols); i += 50 {
		var j int = i + 50
		if j > len(symbols)-1 {
			j = len(symbols) - 1
		}
		h.Repository.Flush(symbols[i:j])
	}

	return nil
}
