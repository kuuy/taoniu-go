package spiders

import (
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spidersRepositories "taoniu.local/cryptos/repositories/binance/currencies/spiders"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type CrawlsHandler struct {
  Db                *gorm.DB
  CrawlsRepository  *spidersRepositories.CrawlsRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewCrawlsCommand() *cli.Command {
  var h CrawlsHandler
  return &cli.Command{
    Name:  "crawls",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CrawlsHandler{
        Db: common.NewDB(1),
      }
      h.CrawlsRepository = &spidersRepositories.CrawlsRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "request",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Request(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *CrawlsHandler) Request() error {
  log.Println("crawl request processing...")
  for _, asset := range h.SymbolsRepository.Currencies() {
    err := h.CrawlsRepository.Request(asset)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}
