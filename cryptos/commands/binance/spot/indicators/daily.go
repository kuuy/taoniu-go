package indicators

import (
	"context"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"log"
	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type DailyHandler struct {
	Db         *gorm.DB
	Repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
	var h DailyHandler
	return &cli.Command{
		Name:  "daily",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = DailyHandler{
				Db: pool.NewDB(),
			}
			h.Repository = &repositories.DailyRepository{
				Db:  h.Db,
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "pivot",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.pivot(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "atr",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.atr(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "zlema",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.zlema(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "ha_zlema",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.haZlema(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "kdj",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.kdj(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "bbands",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.bBands(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DailyHandler) atr() error {
	log.Println("daily atr processing...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		h.Repository.Atr(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) zlema() error {
	log.Println("daily zlema processing...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		h.Repository.Zlema(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) haZlema() error {
	log.Println("daily ha_zlema processing...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		h.Repository.HaZlema(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) kdj() error {
	log.Println("daily kdj indicator...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		h.Repository.Kdj(symbol, 9, 3, 100)
	}
	return nil
}

func (h *DailyHandler) bBands() error {
	log.Println("daily boll bands indicator...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		h.Repository.BBands(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) pivot() error {
	log.Println("daily pivot indicator...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		h.Repository.Pivot(symbol)
	}
	return nil
}
