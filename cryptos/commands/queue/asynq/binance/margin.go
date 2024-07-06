package binance

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/queue/asynq/binance/margin"
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
