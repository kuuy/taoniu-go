package tasks

import (
  "context"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type FundingRateHandler struct {
  Rdb                   *redis.Client
  Ctx                   context.Context
  FundingRateRepository *repositories.FundingRateRepository
  SymbolsRepository     *repositories.SymbolsRepository
  ScalpingRepository    *repositories.ScalpingRepository
}

func NewFundingRateCommand() *cli.Command {
  var h FundingRateHandler
  return &cli.Command{
    Name:  "funding-rate",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FundingRateHandler{
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.FundingRateRepository = &repositories.FundingRateRepository{
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
          return h.Flush()
        },
      },
    },
  }
}

func (h *FundingRateHandler) Flush() (err error) {
  log.Printf("flushing futures funding rate...")
  err = h.FundingRateRepository.Flush()
  time.Sleep(5 * time.Second)
  return
}
