package swap

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/raydium/swap"
)

type SymbolsHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
  var h SymbolsHandler
  return &cli.Command{
    Name:  "symbols",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SymbolsHandler{
        Db:  common.NewDB(3),
        Rdb: common.NewRedis(3),
        Ctx: context.Background(),
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "apply",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol can not be empty")
            return nil
          }
          baseAddress := c.Args().Get(1)
          if baseAddress == "" {
            log.Fatal("base address not be empty")
            return nil
          }
          quoteAddress := c.Args().Get(2)
          if quoteAddress == "" {
            log.Fatal("quote address not be empty")
            return nil
          }
          if err := h.Apply(symbol, baseAddress, quoteAddress); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SymbolsHandler) Apply(symbol, baseAddress, quoteAddress string) error {
  log.Println("raydium swap symbols apply...")

  err := h.SymbolsRepository.Apply(symbol, baseAddress, quoteAddress)
  if err != nil {
    return err
  }

  return nil
}
