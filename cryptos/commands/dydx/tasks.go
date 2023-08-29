package dydx

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/dydx/tasks"
)

func NewTasksCommand() *cli.Command {
  return &cli.Command{
    Name:  "tasks",
    Usage: "",
    Subcommands: []*cli.Command{
      tasks.NewKlinesCommand(),
    },
  }
}
