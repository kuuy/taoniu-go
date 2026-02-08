package futures

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/tasks"
)

func NewTasksCommand() *cli.Command {
  return &cli.Command{
    Name:  "tasks",
    Usage: "",
    Subcommands: []*cli.Command{
      tasks.NewAccountCommand(),
      tasks.NewSymbolsCommand(),
      tasks.NewKlinesCommand(),
      tasks.NewIndicatorsCommand(),
      tasks.NewStrategiesCommand(),
      tasks.NewPlansCommand(),
      tasks.NewScalpingCommand(),
      tasks.NewTradingsCommand(),
      tasks.NewAnalysisCommand(),
    },
  }
}
