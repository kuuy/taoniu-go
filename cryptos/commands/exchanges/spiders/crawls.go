package spiders

import (
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  pool "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/exchanges/spiders"
)

type CrawlsHandler struct {
  Db         *gorm.DB
  Repository *repositories.CrawlsRepository
}

func NewCrawlsCommand() *cli.Command {
  var h CrawlsHandler
  return &cli.Command{
    Name:  "crawls",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CrawlsHandler{
        Db: pool.NewDB(1),
      }
      h.Repository = &repositories.CrawlsRepository{
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
  return h.Repository.Request()
}
