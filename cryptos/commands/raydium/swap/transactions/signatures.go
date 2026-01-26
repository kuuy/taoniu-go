package transactions

import (
  "context"
  "log"
  "strconv"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/raydium/swap/transactions"
)

type SignaturesHandler struct {
  Db                   *gorm.DB
  Ctx                  context.Context
  SignaturesRepository *repositories.SignaturesRepository
}

func NewSignaturesCommand() *cli.Command {
  var h SignaturesHandler
  return &cli.Command{
    Name:  "signatures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SignaturesHandler{
        Db:  common.NewDB(3),
        Ctx: context.Background(),
      }
      h.SignaturesRepository = &repositories.SignaturesRepository{
        Db:  h.Db,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          limit, _ := strconv.Atoi(c.Args().Get(0))
          if limit < 1 || limit > 1000 {
            log.Fatal("limit not in 1~1000")
            return nil
          }
          if err := h.SignaturesRepository.Flush(limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SignaturesHandler) Flush(limit int) (err error) {
  log.Println("raydium swap transactions signatures flush...", limit)
  err = h.SignaturesRepository.Flush(limit)
  return
}
