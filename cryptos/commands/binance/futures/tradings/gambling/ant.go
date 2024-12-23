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
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings/gambling"
)

type AntHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AntRepository
}

func NewAntCommand() *cli.Command {
  var h AntHandler
  return &cli.Command{
    Name:  "ant",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AntHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AntRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &futuresRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.OrdersRepository = &futuresRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.PositionRepository = &futuresRepositories.PositionsRepository{
        Db: h.Db,
      }
      h.Repository.GamblingRepository = &futuresRepositories.GamblingRepository{}
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
        Name:  "take",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Take(); err != nil {
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
  log.Println("futures tradings gambling ant place...")
  ids := h.Repository.PlaceIds()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_PLACE, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.Repository.Place(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}

func (h *AntHandler) Take() error {
  log.Println("futures tradings gambling ant take...")
  ids := h.Repository.TakeIds()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_TAKE, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.Repository.Take(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}

func (h *AntHandler) Flush() error {
  log.Println("futures tradings gambling ant flush...")
  ids := h.Repository.AntIds()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_FLUSH, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.Repository.Flush(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}
