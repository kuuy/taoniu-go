package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type TradingsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.TradingsRepository
}

func NewTradingsCommand() *cli.Command {
	h := TradingsHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
		Repository: &repositories.TradingsRepository{
			Db:  pool.NewDB(),
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "tradings",
		Usage: "",
		Subcommands: []*cli.Command{
			{
				Name:  "scalping",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.scalping(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "buy",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.buy(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *TradingsHandler) scalping() error {
	log.Println("spot margin isolated tradings scalping...")
	h.Repository.Scalping()
	return nil
}

func (h *TradingsHandler) buy() error {
	symbol := "AVAXBUSD"
	price := 15.427547306193494
	orderId, err := h.Repository.Buy(symbol, price, 10)
	if err != nil {
		return err
	}
	log.Println("order:", symbol, orderId)

	return nil
}
