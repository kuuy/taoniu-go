package futures

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/patterns"
)

func NewPatternsCommand() *cli.Command {
  return &cli.Command{
    Name:  "patterns",
    Usage: "",
    Subcommands: []*cli.Command{
      patterns.NewCandlesticksCommand(),
    },
  }
}
