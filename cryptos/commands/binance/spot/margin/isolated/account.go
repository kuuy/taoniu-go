package isolated

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type AccountHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AccountRepository
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
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.Repository.TradingsRepository.FishersRepository = &tradingsRepositories.FishersRepository{
        Db: h.Db,
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
      {
        Name:  "transfer",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Transfer(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "loan",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Loan(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "repay",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Repay(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "collect",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Collect(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "liquidate",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Liquidate(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *AccountHandler) Flush() error {
  log.Println("margin isolated account flush processing...")
  return h.Repository.Flush()
}

func (h *AccountHandler) Transfer() error {
  log.Println("margin isolated account transfer...")
  asset := "AAVE"
  symbol := "AAVEUSDT"
  quantity := 0.01
  from := "ISOLATED_MARGIN"
  to := "SPOT"
  transferId, err := h.Repository.Transfer(asset, symbol, from, to, quantity)
  if err != nil {
    return err
  }
  log.Println("transferId", transferId)
  return nil
}

func (h *AccountHandler) Loan() error {
  asset := "USDT"
  symbol := "ATOMUSDT"
  amount := 0.01
  transferId, err := h.Repository.Loan(asset, symbol, amount, true)
  if err != nil {
    return err
  }
  log.Println("transferId", transferId)
  return nil
}

func (h *AccountHandler) Repay() error {
  asset := "USDT"
  symbol := "ATOMUSDT"
  amount := 0.01
  transferId, err := h.Repository.Repay(asset, symbol, amount, true)
  if err != nil {
    return err
  }
  log.Println("transferId", transferId)
  return nil
}

func (h *AccountHandler) Collect() error {
  log.Println("margin isolated account collect...")
  return h.Repository.Collect()
}

func (h *AccountHandler) Liquidate() error {
  log.Println("margin isolated account liquidate...")
  return h.Repository.Liquidate()
}
