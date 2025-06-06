package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance"
)

func NewBinanceCommand() *cli.Command {
  return &cli.Command{
    Name:  "binance",
    Usage: "",
    Subcommands: []*cli.Command{
      binance.NewAccountCommand(),
      binance.NewSpotCommand(),
      binance.NewMarginCommand(),
      binance.NewFutoresCommand(),
      binance.NewSavingsCommand(),
      binance.NewServerCommand(),
    },
  }
}
