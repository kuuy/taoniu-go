package binance

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/cron/binance/margin"
)

func NewMarginCommand() *cli.Command {
  return &cli.Command{
    Name:  "margin",
    Usage: "",
    Subcommands: []*cli.Command{
      margin.NewCrossCommand(),
    },
  }
}
