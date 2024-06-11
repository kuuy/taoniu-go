package proxies

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/proxies"
)

type SourcesHandler struct {
  Repository *repositories.SourcesRepository
}

func NewSourcesCommand() *cli.Command {
  var h SourcesHandler
  return &cli.Command{
    Name:  "Sources",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SourcesHandler{
        Repository: &repositories.SourcesRepository{
          Db:  common.NewDB(1),
          Rdb: common.NewRedis(1),
          Ctx: context.Background(),
        },
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "add",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.add(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SourcesHandler) add() error {
  log.Println("sources add processing...")
  h.Repository.Add()
  return nil
}
