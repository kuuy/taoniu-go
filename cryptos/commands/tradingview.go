package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/tradingview"
)

func NewTradingviewCommand() *cli.Command {
  return &cli.Command{
    Name:  "tradingview",
    Usage: "",
    Subcommands: []*cli.Command{
      tradingview.NewAnalysisCommand(),
      tradingview.NewScannerCommand(),
    },
  }
}
