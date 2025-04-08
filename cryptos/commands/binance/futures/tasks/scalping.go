package tasks

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/tasks/scalping"
)

func NewScalpingCommand() *cli.Command {
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Subcommands: []*cli.Command{
      scalping.NewPlansCommand(),
    },
  }
}
