package spot

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/commands/binance/spot/tradings"
  "taoniu.local/cryptos/common"
  savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type TradingsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.TradingsRepository
}

func NewTradingsCommand() *cli.Command {
  var h TradingsHandler
  return &cli.Command{
    Name:  "tradings",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TradingsHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.Repository.AccountRepository = &repositories.AccountRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.ProductsRepository = &savingsRepositories.ProductsRepository{
        Db: h.Db,
      }
      h.Repository.FishersRepository = &tradingsRepositories.FishersRepository{
        Db: h.Db,
      }
      h.Repository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      h.Repository.TriggersRepository = &tradingsRepositories.TriggersRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "pending",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.pending(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "collect",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.collect(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      tradings.NewFishersCommand(),
      tradings.NewScalpingCommand(),
      tradings.NewTriggersCommand(),
    },
  }
}

func (h *TradingsHandler) pending() error {
  log.Println("spot tradings pending...")
  data := h.Repository.Pending()
  log.Println(data)
  return nil
}

func (h *TradingsHandler) collect() error {
  log.Println("spot tradings collect...")
  return h.Repository.Collect()
}
