package tradings

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type ScalpingHandler struct {
	Db              *gorm.DB
	Rdb             *redis.Client
	Ctx             context.Context
	Repository      *repositories.ScalpingRepository
	PlansRepository *plansRepositories.DailyRepository
}

func NewScalpingCommand() *cli.Command {
	var h ScalpingHandler
	return &cli.Command{
		Name:  "scalping",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = ScalpingHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.ScalpingRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.AccountRepository = &spotRepositories.AccountRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.PlansRepository = &plansRepositories.DailyRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
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

func (h *ScalpingHandler) Place() error {
	log.Println("spot tradings scalping place...")
	plan, err := h.PlansRepository.Filter()
	if err != nil {
		return err
	}
	h.Repository.Place(plan)
	return nil
}

func (h *ScalpingHandler) Flush() error {
	log.Println("spot tradings scalping flush...")
	symbols := h.Repository.Scan()
	log.Println("symbols", symbols)
	for _, symbol := range symbols {
		err := h.Repository.Flush(symbol)
		if err != nil {
			log.Println("error", err)
		}
	}
	return nil
}
