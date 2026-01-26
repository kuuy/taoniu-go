package swap

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"

  "taoniu.local/cryptos/commands/raydium/swap/transactions"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/raydium/swap"
)

type TransactionsHandler struct {
  Db                     *gorm.DB
  Rdb                    *redis.Client
  Ctx                    context.Context
  TransactionsRepository *repositories.TransactionsRepository
}

func NewTransactionsCommand() *cli.Command {
  var h TransactionsHandler
  return &cli.Command{
    Name:  "transactions",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TransactionsHandler{
        Db:  common.NewDB(3),
        Rdb: common.NewRedis(3),
        Ctx: context.Background(),
      }
      h.TransactionsRepository = &repositories.TransactionsRepository{
        Db:  h.Db,
        Ctx: h.Ctx,
      }
      h.TransactionsRepository.MintsRepository = &repositories.MintsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      transactions.NewSignaturesCommand(),
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if c.Args().Get(0) == "" {
            log.Fatal("signature is empty")
            return nil
          }
          signature := c.Args().Get(0)
          if err := h.Flush(signature); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TransactionsHandler) Flush(signature string) (err error) {
  log.Println("raydium swap transactions flush...", signature)
  err = h.TransactionsRepository.Flush(signature)
  return
}
