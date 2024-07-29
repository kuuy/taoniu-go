package spot

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/spot/tasks"
)

func NewTasksCommand() *cli.Command {
  return &cli.Command{
    Name:  "tasks",
    Usage: "",
    Subcommands: []*cli.Command{
      tasks.NewSymbolsCommand(),
      tasks.NewKlinesCommand(),
      tasks.NewStrategiesCommand(),
      tasks.NewPlansCommand(),
      tasks.NewTradingsCommand(),
      tasks.NewAnalysisCommand(),
    },
  }
}
