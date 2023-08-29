package plans

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/plans"
)

type MinutelyHandler struct {
  Repository *repositories.MinutelyRepository
}

func NewMinutelyCommand() *cli.Command {
  var h MinutelyHandler
  return &cli.Command{
    Name:  "minutely",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MinutelyHandler{}
      h.Repository = &repositories.MinutelyRepository{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.fix(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *MinutelyHandler) flush() error {
  log.Println("spot plans daily flush...")
  return h.Repository.Flush()
}

func (h *MinutelyHandler) fix() error {
  log.Println("spot plans daily fix...")
  return h.Repository.Fix()
}
