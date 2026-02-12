package futures

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type FundingRateHandler struct {
  Db                    *gorm.DB
  Rdb                   *redis.Client
  Ctx                   context.Context
  Nats                  *nats.Conn
  FundingRateRepository *repositories.FundingRateRepository
}

func NewFundingRateCommand() *cli.Command {
  var h FundingRateHandler
  return &cli.Command{
    Name:  "funding-rate",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FundingRateHandler{
        Db:   common.NewDB(2),
        Rdb:  common.NewRedis(2),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.FundingRateRepository = &repositories.FundingRateRepository{
        Db:  h.Db,
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

func (h *FundingRateHandler) Flush() error {
  log.Println("binance futures funding rate flush...")
  err := h.FundingRateRepository.Flush()
  if err != nil {
    log.Println("kline flush error", err)
  }
  return nil
}
