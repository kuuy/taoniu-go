package binance

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/savings"
)

func NewSavingsCommand() *cli.Command {
  return &cli.Command{
    Name:  "savings",
    Usage: "",
    Subcommands: []*cli.Command{
      savings.NewAccountCommand(),
      savings.NewProductsCommand(),
    },
  }
}
