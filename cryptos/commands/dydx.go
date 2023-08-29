package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/dydx"
)

func NewDydxCommand() *cli.Command {
  return &cli.Command{
    Name:  "dydx",
    Usage: "",
    Subcommands: []*cli.Command{
      dydx.NewServerCommand(),
      dydx.NewMarketsCommand(),
      dydx.NewOrderbookCommand(),
      dydx.NewKlinesCommand(),
      dydx.NewIndicatorsCommand(),
      dydx.NewStrategiesCommand(),
      dydx.NewPlansCommand(),
      dydx.NewAccountCommand(),
      dydx.NewPositionsCommand(),
      dydx.NewOrdersCommand(),
      dydx.NewScalpingCommand(),
      dydx.NewTriggersCommand(),
      dydx.NewTradingsCommand(),
      dydx.NewStreamCommand(),
      dydx.NewTasksCommand(),
    },
  }
}
