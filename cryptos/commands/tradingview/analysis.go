package tradingview

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"log"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/tradingview"
)

type AnalysisHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.AnalysisRepository
}

func NewAnalysisCommand() *cli.Command {
	var h AnalysisHandler
	return &cli.Command{
		Name:  "analysis",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = AnalysisHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.AnalysisRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
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
				Name:  "summary",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Summary(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *AnalysisHandler) Flush() error {
	log.Println("analysis flush processing...")
	err := h.Repository.Flush("BINANCE", "AVAXBUSD", "1m")
	if err != nil {
		return err
	}
	return nil
}

func (h *AnalysisHandler) Summary() error {
	summary, err := h.Repository.Summary("BINANCE", "AVAXBUSD", "1m")
	if err != nil {
		return err
	}
	log.Println("summary", summary)
	return nil
}
