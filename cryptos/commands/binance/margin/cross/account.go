package cross

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
  spotTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
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
        Db:   common.NewDB(1),
        Rdb:  common.NewRedis(1),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.Repository = &repositories.AccountRepository{
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Nats: h.Nats,
      }
      h.SymbolsRepository = &marginRepositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SpotAccountRepository = &spotRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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
      {
        Name:  "borrow",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Borrow(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *AccountHandler) Flush() error {
  log.Println("binance margin cross account flush processing...")
  return h.Repository.Flush()
}

func (h *AccountHandler) Balance() error {
  log.Println("binance margin cross account flush processing...")
  balance, err := h.Repository.Balance("USDT")
  if err != nil {
    return err
  }
  log.Println("balance", balance)
  return nil
}

func (h *AccountHandler) Collect() error {
  log.Println("binance margin cross account collect processing...")
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
  for _, currency := range h.SymbolsRepository.Currencies() {
    balance, err = h.SpotAccountRepository.Balance(currency)
    if err != nil {
      continue
    }
    if balance["free"] <= 0.0 {
      continue
    }
    transferId, err = h.Repository.Transfer(currency, 1, balance["free"])
    if err != nil {
      log.Println("collect error", currency, err.Error())
    } else {
      log.Println("collect asset", currency, balance["free"], transferId)
    }
  }

  h.Repository.Flush()
  h.SpotAccountRepository.Flush()
  return nil
}

func (h *AccountHandler) Liquidate() (err error) {
  log.Println("binance margin cross account liquidate processing...")
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
  for _, asset := range h.SymbolsRepository.Currencies() {
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

func (h *AccountHandler) Borrow() (err error) {
  log.Println("binance margin cross account borrow processing...")
  transferId, err := h.Repository.Borrow("USDT", 10.5)
  log.Println("transferId", transferId)
  return
}
