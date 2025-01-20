package binance

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures"
)

func NewFutoresCommand() *cli.Command {
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Subcommands: []*cli.Command{
      futures.NewSymbolsCommand(),
      futures.NewTickersCommand(),
      futures.NewKlinesCommand(),
      futures.NewDepthCommand(),
      futures.NewIndicatorsCommand(),
      futures.NewStrategiesCommand(),
      futures.NewPlansCommand(),
      futures.NewPatternsCommand(),
      futures.NewAccountCommand(),
      futures.NewOrdersCommand(),
      futures.NewPositionsCommand(),
      futures.NewScalpingCommand(),
      futures.NewTradingsCommand(),
      futures.NewStreamCommand(),
      futures.NewTasksCommand(),
      futures.NewAnalysisCommand(),
      futures.NewGamblingCommand(),
    },
  }
}
