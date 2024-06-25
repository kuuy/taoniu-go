package cross

import (
  "context"
  "gorm.io/gorm"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  spotTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type AccountHandler struct {
  Db                         *gorm.DB
  Rdb                        *redis.Client
  Ctx                        context.Context
  Repository                 *repositories.AccountRepository
  SymbolsRepository          *marginRepositories.SymbolsRepository
  SpotAccountRepository      *spotRepositories.AccountRepository
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
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.SymbolsRepository = &marginRepositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SpotAccountRepository = &spotRepositories.AccountRepository{
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

func (h *AccountHandler) Collect() error {
  log.Println("cross margin account collect processing...")
  var transferId int64
  balance, err := h.SpotAccountRepository.Balance("USDT")
  if err == nil {
    if balance["free"] > 0.0 {
      h.Repository.Transfer("USDT", 1, balance["free"])
      transferId, err = h.Repository.Transfer("USDT", 1, balance["free"])
      if err != nil {
        log.Println("collect error", "USDT", err.Error())
      } else {
        log.Println("collect asset", "USDT", balance["free"], transferId)
      }
    }
  }
  for _, asset := range h.SymbolsRepository.Assets() {
    balance, err = h.SpotAccountRepository.Balance(asset)
    if err != nil {
      continue
    }
    if balance["free"] <= 0.0 {
      continue
    }
    transferId, err = h.Repository.Transfer(asset, 1, balance["free"])
    if err != nil {
      log.Println("collect error", asset, err.Error())
    } else {
      log.Println("collect asset", asset, balance["free"], transferId)
    }
  }
  h.SpotAccountRepository.Flush()
  h.Repository.Flush()
  return nil
}

func (h *AccountHandler) Liquidate() (err error) {
  log.Println("cross margin account liquidate processing...")
  var transferId int64
  balance, err := h.Repository.Balance("USDT")
  if err == nil {
    if balance["free"] > 0.0 {
      transferId, err = h.Repository.Transfer("USDT", 2, balance["free"])
      if err != nil {
        log.Println("liquidate error", "USDT", err.Error())
      } else {
        log.Println("liquidate asset", "USDT", balance["free"], transferId)
      }
    }
  }
  for _, asset := range h.SymbolsRepository.Assets() {
    balance, err = h.Repository.Balance(asset)
    if err != nil {
      continue
    }
    if balance["free"] <= 0.0 {
      continue
    }
    transferId, err = h.Repository.Transfer(asset, 2, balance["free"])
    if err != nil {
      log.Println("liquidate error", asset, err.Error())
    } else {
      log.Println("liquidate asset", asset, balance["free"], transferId)
    }
  }
  h.SpotAccountRepository.Flush()
  h.Repository.Flush()
  return
}
