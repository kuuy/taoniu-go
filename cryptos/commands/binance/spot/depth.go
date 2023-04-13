package spot

import (
	"context"
	"log"
	tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type symbolsRepository interface {
	Scan() []string
}

type DepthHandler struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	Repository        *repositories.DepthRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func NewDepthCommand() *cli.Command {
	var h DepthHandler
	return &cli.Command{
		Name:  "depth",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = DepthHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.DepthRepository{
				Db: h.Db,
			}
			h.SymbolsRepository = &repositories.SymbolsRepository{
				Db: h.Db,
			}
			h.SymbolsRepository.TradingsRepository = &repositories.TradingsRepository{
				Db: h.Db,
			}
			h.SymbolsRepository.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					h.Repository.UseProxy = c.Bool("proxy")
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DepthHandler) flush() error {
	log.Println("symbols depth flush...")
	symbols := h.SymbolsRepository.Scan()
	for _, symbol := range symbols {
		err := h.Repository.Flush(symbol)
		if err != nil {
			log.Println("error", err)
		}
	}
	return nil
}
