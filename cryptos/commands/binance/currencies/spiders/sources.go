package spiders

import (
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spidersRepositories "taoniu.local/cryptos/repositories/binance/currencies/spiders"
)

type SourcesHandler struct {
  Db                *gorm.DB
  SourcesRepository *spidersRepositories.SourcesRepository
}

func NewSourcesCommand() *cli.Command {
  var h SourcesHandler
  return &cli.Command{
    Name:  "sources",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SourcesHandler{
        Db: common.NewDB(1),
      }
      h.SourcesRepository = &spidersRepositories.SourcesRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "add",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Add(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SourcesHandler) Add() error {
  log.Println("sources add processing...")
  return h.SourcesRepository.Add()
}
