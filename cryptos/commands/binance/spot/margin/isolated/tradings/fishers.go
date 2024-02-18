package tradings

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/spot/margin/isolated/tradings/fishers"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  tvRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type FishersHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.FishersRepository
}

func NewFishersCommand() *cli.Command {
  var h FishersHandler
  return &cli.Command{
    Name:  "fishers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FishersHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.FishersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AnalysisRepository = &tvRepositories.AnalysisRepository{
        Db: h.Db,
      }
      marginRepository := &spotRepositories.MarginRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = marginRepository.Isolated().Account()
      h.Repository.OrdersRepository = marginRepository.Orders()
      return nil
    },
    Subcommands: []*cli.Command{
      fishers.NewGridsCommand(),
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

func (h *FishersHandler) Flush() error {
  symbols := h.Repository.Scan()
  for _, symbol := range symbols {
    err := h.Repository.Flush(symbol)
    if err != nil {
      log.Println("fishers flush error", err)
    }
  }
  return nil
}

func (h *FishersHandler) Place() error {
  symbols := h.Repository.Scan()
  for _, symbol := range symbols {
    err := h.Repository.Place(symbol)
    if err != nil {
      log.Println("fishers place error", err)
    }
  }
  return nil
}
