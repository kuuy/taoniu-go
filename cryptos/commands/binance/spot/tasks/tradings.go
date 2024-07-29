package tasks

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/spot/tasks/tradings"
)

func NewTradingsCommand() *cli.Command {
  return &cli.Command{
    Name:  "tradings",
    Usage: "",
    Subcommands: []*cli.Command{
      tradings.NewTriggersCommand(),
      tradings.NewScalpingCommand(),
    },
  }
}
