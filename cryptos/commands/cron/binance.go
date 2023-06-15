package cron

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/cron/binance"
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
