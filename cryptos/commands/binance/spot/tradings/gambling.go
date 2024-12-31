package tradings

import (
  "context"
  "gorm.io/gorm"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/commands/binance/spot/tradings/gambling"
)

type GamblingHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewGamblingCommand() *cli.Command {
  return &cli.Command{
    Name:  "gambling",
    Usage: "",
    Subcommands: []*cli.Command{
      gambling.NewScalpingCommand(),
      gambling.NewAntCommand(),
    },
  }
}
