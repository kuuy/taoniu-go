package futures

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type AccountHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Nats              *nats.Conn
  AccountRepository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Db:   common.NewDB(2),
        Rdb:  common.NewRedis(2),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.AccountRepository = &repositories.AccountRepository{
        Db:   h.Db,
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Nats: h.Nats,
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
  log.Println("binance futures account flush processing...")
  return h.AccountRepository.Flush()
}
