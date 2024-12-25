package gambling

import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings/gambling"
)

type ScalpingHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Repository         *repositories.ScalpingRepository
  TradingsRepository *tradingsRepositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.ScalpingRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &spotRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.PositionRepository = &spotRepositories.PositionsRepository{
        Db: h.Db,
      }
      h.TradingsRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "place",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Place(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Place() error {
  ids := h.TradingsRepository.ScalpingIds()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_PLACE, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }

    err := h.Repository.Place(id)
    if err != nil {
      log.Println("gambling scalping place error", err)
    }

    mutex.Unlock()
  }
  return nil
}
