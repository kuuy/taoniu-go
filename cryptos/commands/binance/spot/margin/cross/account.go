package cross

import (
  "context"
  "gorm.io/gorm"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/common"

  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  spotTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type AccountHandler struct {
  Db                         *gorm.DB
  Rdb                        *redis.Client
  Ctx                        context.Context
  Repository                 *repositories.AccountRepository
  SpotSymbolsRepository      *spotRepositories.SymbolsRepository
  SpotTradingsRepository     *spotRepositories.TradingsRepository
  IsolatedAccountRepository  *isolatedRepositories.AccountRepository
  IsolatedTradingsRepository *isolatedRepositories.TradingsRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.SpotSymbolsRepository = &spotRepositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SpotTradingsRepository = &spotRepositories.TradingsRepository{
        Db: h.Db,
      }
      h.SpotTradingsRepository.ScalpingRepository = &spotTradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      h.SpotTradingsRepository.TriggersRepository = &spotTradingsRepositories.TriggersRepository{
        Db: h.Db,
      }
      //h.IsolatedAccountRepository = &isolatedRepositories.AccountRepository{}
      //h.IsolatedTradingsRepository = &isolatedRepositories.TradingsRepository{
      //  Db: h.Db,
      //}
      //h.IsolatedTradingsRepository.FishersRepository = &isolatedTradingsRepositories.FishersRepository{
      //  Db: h.Db,
      //}
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
        Name:  "balance",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Balance(); err != nil {
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
  log.Println("cross margin account flush processing...")
  return h.Repository.Flush()
}

func (h *AccountHandler) Balance() error {
  log.Println("cross margin account flush processing...")
  balance, err := h.Repository.Balance("USDT")
  if err != nil {
    return err
  }
  log.Println("balance", balance)
  return nil
}

func (h *AccountHandler) Transfer() error {
  log.Println("cross margin account transfer processing...")
  //return h.Repository.Flush()
  return nil
}

func (h *AccountHandler) Collect() error {
  log.Println("cross margin account collect processing...")
  //return h.Repository.Flush()
  return nil
}

func (h *AccountHandler) Liquidate() (err error) {
  log.Println("cross margin account liquidate processing...")
  for _, symbol := range h.IsolatedTradingsRepository.Scan() {
    entity, err := h.SpotSymbolsRepository.Get(symbol)
    if err != nil {
      continue
    }
    balance, err := h.Repository.Balance(entity.BaseAsset)
    if err != nil {
      log.Println("error", symbol, err.Error())
      continue
    }
    if balance["free"] <= 0.0 {
      continue
    }
    var transferId int64
    transferId, err = h.Repository.Transfer(entity.BaseAsset, 2, balance["free"])
    if err != nil {
      log.Println("transfer error", symbol, err.Error())
      continue
    }
    log.Println("transfer to spot", symbol, transferId)
    transferId, err = h.IsolatedAccountRepository.Transfer(entity.BaseAsset, symbol, "SPOT", "ISOLATED_MARGIN", balance["free"])
    if err != nil {
      log.Println("transfer error", symbol, err.Error())
      continue
    }
    h.IsolatedAccountRepository.Transfer(entity.QuoteAsset, symbol, "SPOT", "ISOLATED_MARGIN", 0.5)
    log.Println("transfer to margin isolated", symbol, transferId)
  }
  err = h.Repository.Flush()
  if err != nil {
    return
  }
  for _, symbol := range h.SpotTradingsRepository.Scan() {
    entity, err := h.SpotSymbolsRepository.Get(symbol)
    if err != nil {
      continue
    }
    balance, err := h.Repository.Balance(entity.BaseAsset)
    if err != nil {
      log.Println("error", symbol, err.Error())
      continue
    }
    if balance["free"] <= 0.0 {
      continue
    }
    transferId, err := h.Repository.Transfer(entity.BaseAsset, 2, balance["free"])
    if err != nil {
      log.Println("transfer error", symbol, err.Error())
      continue
    }
    log.Println("transfer to spot", symbol, transferId)
  }
  //return h.Repository.Flush()
  return
}
