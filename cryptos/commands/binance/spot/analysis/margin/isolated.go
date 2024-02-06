package margin

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/commands/binance/spot/analysis/margin/isolated"
)

type IsolatedHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewIsolatedCommand() *cli.Command {
  return &cli.Command{
    Name:  "isolated",
    Usage: "",
    Subcommands: []*cli.Command{
      isolated.NewTradingsCommand(),
    },
  }
}
