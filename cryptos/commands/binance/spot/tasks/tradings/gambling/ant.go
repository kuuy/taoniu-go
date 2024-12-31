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
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  gamblingTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings/gambling"
)

type AntHandler struct {
  Db                         *gorm.DB
  Rdb                        *redis.Client
  Ctx                        context.Context
  GamblingTradingsRepository *gamblingTradingsRepositories.AntRepository
}

func NewAntCommand() *cli.Command {
  var h AntHandler
  return &cli.Command{
    Name:  "ant",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AntHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.GamblingTradingsRepository = &gamblingTradingsRepositories.AntRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.SymbolsRepository = &repositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.AccountRepository = &repositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.OrdersRepository = &repositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.GamblingTradingsRepository.PositionRepository = &repositories.PositionsRepository{
        Db: h.Db,
      }
      h.GamblingTradingsRepository.GamblingRepository = &repositories.GamblingRepository{}
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

func (h *AntHandler) Place() error {
  log.Println("spot tradings gambling ant place...")
  ids := h.GamblingTradingsRepository.Ids()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_PLACE, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.GamblingTradingsRepository.Place(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}

func (h *AntHandler) Flush() error {
  log.Println("spot tradings gambling ant flush...")
  ids := h.GamblingTradingsRepository.Ids()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_FLUSH, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.GamblingTradingsRepository.Flush(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}
