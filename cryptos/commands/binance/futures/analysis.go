package futures

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/analysis"
)

func NewAnalysisCommand() *cli.Command {
  return &cli.Command{
    Name:  "analysis",
    Usage: "",
    Subcommands: []*cli.Command{
      analysis.NewTradingsCommand(),
    },
  }
}
