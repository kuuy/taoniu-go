package cross

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/margin/cross/tasks"
)

func NewTasksCommand() *cli.Command {
  return &cli.Command{
    Name:  "tasks",
    Usage: "",
    Subcommands: []*cli.Command{
      tasks.NewAccountCommand(),
      tasks.NewTradingsCommand(),
    },
  }
}
