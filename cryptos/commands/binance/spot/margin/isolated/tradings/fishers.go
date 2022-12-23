package tradings

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type FishersHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.FishersRepository
}

func NewFishersCommand() *cli.Command {
	var h FishersHandler
	return &cli.Command{
		Name:  "fishers",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = FishersHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.FishersRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			marginRepository := &spotRepositories.MarginRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.AccountRepository = marginRepository.Isolated().Account()
			h.Repository.OrdersRepository = marginRepository.Orders()
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "apply",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Apply(); err != nil {
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
				Name:  "place",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Place(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *FishersHandler) Flush() error {
	symbols := h.Repository.Scan()
	log.Println("symbols", symbols)
	return nil
}

func (h *FishersHandler) Place() error {
	symbols := h.Repository.Scan()
	for _, symbol := range symbols {
		err := h.Repository.Place(symbol)
		if err != nil {
			log.Println("fishers place error", err)
		}
	}
	return nil
}

func (h *FishersHandler) Apply() error {
	symbol := "AVAXBUSD"
	amount := 10.0
	balance := 360.0
	targetBalance := 400.0
	stopBalance := 100.0
	tickers := make([][]float64, 3)
	tickers[0] = []float64{13.33, 13.1, 12.60, 12.52, 12.25}
	tickers[1] = []float64{11.8, 11.72, 11.61, 11.57, 11.55}
	tickers[2] = []float64{11.42, 11.35, 11.28}
	return h.Repository.Apply(
		symbol,
		amount,
		balance,
		targetBalance,
		stopBalance,
		tickers,
	)
}
