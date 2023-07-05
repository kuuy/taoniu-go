package asynq

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/queue/asynq/binance"
)

func NewBinanceCommand() *cli.Command {
  return &cli.Command{
    Name:  "binance",
    Usage: "",
    Subcommands: []*cli.Command{
      binance.NewFuturesCommand(),
      binance.NewSpotCommand(),
    },
  }
}
