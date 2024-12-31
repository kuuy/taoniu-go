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
  config "taoniu.local/cryptos/config/binance/futures"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
  gamblingTradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings/gambling"
)

type ScalpingHandler struct {
  Db                         *gorm.DB
  Rdb                        *redis.Client
  Ctx                        context.Context
  GamblingTradingsRepository *gamblingTradingsRepositories.ScalpingRepository
  TradingsRepository         *tradingsRepositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.GamblingTradingsRepository = &gamblingTradingsRepositories.ScalpingRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.AccountRepository = &futuresRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.OrdersRepository = &futuresRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.PositionRepository = &futuresRepositories.PositionsRepository{
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

    err := h.GamblingTradingsRepository.Place(id)
    if err != nil {
      log.Println("gambling scalping place error", err)
    }

    mutex.Unlock()
  }
  return nil
}
