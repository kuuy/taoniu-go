package binance

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance"
)

type AccountHandler struct {
  Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{}
      h.Repository = &repositories.AccountRepository{
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "balances",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.balances(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *AccountHandler) balances() error {
  log.Println("account balances processing...")
  balances, err := h.Repository.Balances()
  if err != nil {
    return err
  }
  log.Println("balances", balances)
  return nil
}
