package daily

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators/daily"
)

type RankingHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.RankingRepository
}

func NewRankingCommand() *cli.Command {
  var h RankingHandler
  return &cli.Command{
    Name:  "ranking",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = RankingHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.RankingRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "display",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Display(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *RankingHandler) Display() error {
  log.Println("binance futures indicators daily ranking...")
  symbol := ""
  fields := []string{
    "poc",
    "vah",
    "val",
    "poc_ratio",
    "profit_target",
    "stop_loss_point",
    "risk_reward_ratio",
  }
  sortField := "poc_ratio"
  sortType := -1
  current := 1
  pageSize := 50
  result := h.Repository.Listings(symbol, fields, sortField, sortType, current, pageSize)
  log.Println("result", result)

  return nil
}
