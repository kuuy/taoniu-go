package raydium

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/raydium/swap"
)

func NewSwapCommand() *cli.Command {
  return &cli.Command{
    Name:  "swap",
    Usage: "",
    Subcommands: []*cli.Command{
      swap.NewOrdersCommand(),
      swap.NewMintsCommand(),
      swap.NewSymbolsCommand(),
      swap.NewKlinesCommand(),
      swap.NewAccountCommand(),
      swap.NewTransactionsCommand(),
    },
  }
}
