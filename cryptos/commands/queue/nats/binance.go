package nats

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/queue/nats/binance"
)

func NewBinanceCommand() *cli.Command {
  return &cli.Command{
    Name:  "binance",
    Usage: "",
    Subcommands: []*cli.Command{
      binance.NewSpotCommand(),
      binance.NewMarginCommand(),
      binance.NewFuturesCommand(),
    },
  }
}
