package indicators

import (
	"context"
	"log"

	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance/futures"
	repositories "taoniu.local/cryptos/repositories/binance/futures"
	indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type IchimokuCloudHandler struct {
	Repository *repositories.IndicatorsRepository
}

func NewIchimokuCloudCommand() *cli.Command {
	var h IchimokuCloudHandler
	return &cli.Command{
		Name:  "ichimoku-cloud",
		Usage: "",
		Before: func(c *cli.Context) error {
			db := common.NewDB(2)
			rdb := common.NewRedis(2)
			ctx := context.Background()
			h.Repository = &repositories.IndicatorsRepository{
				Db:  db,
				Rdb: rdb,
				Ctx: ctx,
			}
			baseRepository := indicatorsRepositories.BaseRepository{
				Db:  db,
				Rdb: rdb,
				Ctx: ctx,
			}
			h.Repository.IchimokuCloud = &indicatorsRepositories.IchimokuCloudRepository{BaseRepository: baseRepository}
			return nil
		},
		Action: func(c *cli.Context) error {
			symbol := c.Args().Get(1)
			interval := c.Args().Get(0)
			if interval == "" {
				log.Fatal("interval can not be empty")
				return nil
			}
			return h.Flush(symbol, interval)
		},
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "",
				Action: func(c *cli.Context) error {
					symbol := c.Args().Get(1)
					interval := c.Args().Get(0)
					if interval == "" {
						log.Fatal("interval can not be empty")
						return nil
					}
					if err := h.Get(symbol, interval); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					symbol := c.Args().Get(1)
					interval := c.Args().Get(0)
					if interval == "" {
						log.Fatal("interval can not be empty")
						return nil
					}
					if err := h.Flush(symbol, interval); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *IchimokuCloudHandler) Get(symbol string, interval string) (err error) {
	log.Println("indicators ichimoku cloud get...")
	signal, conversionLine, baseLine, senkouSpanA, senkouSpanB, chikouSpan, price, timestamp, err := h.Repository.IchimokuCloud.Get(symbol, interval)
	if err != nil {
		return
	}
	log.Println("result", signal, conversionLine, baseLine, senkouSpanA, senkouSpanB, chikouSpan, price, timestamp)
	return
}

func (h *IchimokuCloudHandler) Flush(symbol string, interval string) error {
	log.Println("indicators ichimoku cloud calc...")
	var symbols []string
	if symbol == "" {
		h.Repository.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
	} else {
		symbols = append(symbols, symbol)
	}
	for _, symbol := range symbols {
		var err error
		switch interval {
		case "1m":
			err = h.Repository.IchimokuCloud.Flush(symbol, interval, 129, 374, 748, 1440)
		case "15m":
			err = h.Repository.IchimokuCloud.Flush(symbol, interval, 60, 174, 349, 672)
		case "4h":
			err = h.Repository.IchimokuCloud.Flush(symbol, interval, 11, 32, 65, 126)
		default:
			err = h.Repository.IchimokuCloud.Flush(symbol, interval, 9, 26, 52, 100)
		}
		if err != nil {
			log.Println("error", err.Error())
		}
	}
	return nil
}
