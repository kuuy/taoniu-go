package tasks

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  marginRepositories "taoniu.local/cryptos/repositories/binance/margin"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross"
  "taoniu.local/cryptos/repositories/binance/margin/isolated"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
)

type AccountHandler struct {
  Db                         *gorm.DB
  Rdb                        *redis.Client
  Ctx                        context.Context
  Nats                       *nats.Conn
  Repository                 *repositories.AccountRepository
  SymbolsRepository          *marginRepositories.SymbolsRepository
  SpotAccountRepository      *spotRepositories.AccountRepository
  SpotTradingsRepository     *spotRepositories.TradingsRepository
  IsolatedAccountRepository  *isolated.AccountRepository
  IsolatedTradingsRepository *isolated.TradingsRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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
  log.Println("binance margin cross tasks account flush processing...")
  return h.Repository.Flush()
}
