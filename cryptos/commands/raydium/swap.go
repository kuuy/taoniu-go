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
      swap.NewIndicatorsCommand(),
      swap.NewStrategiesCommand(),
      swap.NewAccountCommand(),
      swap.NewTransactionsCommand(),
      swap.NewTickersCommand(),
      swap.NewPositionsCommand(),
    },
  }
}
