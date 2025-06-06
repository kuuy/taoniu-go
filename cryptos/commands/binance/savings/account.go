package savings

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
)

type AccountHandler struct {
  SavingsRepository *savingsRepositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{}
      h.SavingsRepository = &savingsRepositories.AccountRepository{
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
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
    },
  }
}

func (h *AccountHandler) Flush() error {
  log.Println("savings account flush...")
  return h.SavingsRepository.Flush()
}
