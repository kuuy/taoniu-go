package spiders

import (
	"log"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/currencies/spiders"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
)

type CrawlsHandler struct {
	Db               *gorm.DB
	Repository       *repositories.CrawlsRepository
	SymbolRepository *spotRepositories.SymbolsRepository
}

func NewCrawlsCommand() *cli.Command {
	var h CrawlsHandler
	return &cli.Command{
		Name:  "crawls",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = CrawlsHandler{
				Db: common.NewDB(),
			}
			h.Repository = &repositories.CrawlsRepository{
				Db: h.Db,
			}
			h.SymbolRepository = &spotRepositories.SymbolsRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "request",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.request(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *CrawlsHandler) request() error {
	log.Println("crawl request processing...")
	for _, currency := range h.SymbolRepository.Currencies() {
		err := h.Repository.Request(currency)
		if err != nil {
			log.Println("error", err)
		}
	}
	return nil
}
