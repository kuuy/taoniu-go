package binance

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/currencies"
)

func NewCurrenciesCommand() *cli.Command {
  return &cli.Command{
    Name:  "currencies",
    Usage: "",
    Subcommands: []*cli.Command{
      currencies.NewSpidersCommand(),
    },
  }
}
