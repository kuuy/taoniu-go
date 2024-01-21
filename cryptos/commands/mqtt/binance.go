package mqtt

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/mqtt/binance"
)

func NewBinanceCommand() *cli.Command {
  return &cli.Command{
    Name:  "binance",
    Usage: "",
    Subcommands: []*cli.Command{
      binance.NewSpotCommand(),
    },
  }
}
