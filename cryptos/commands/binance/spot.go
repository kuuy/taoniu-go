package binance

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/spot"
)

func NewSpotCommand() *cli.Command {
  return &cli.Command{
    Name:  "spot",
    Usage: "",
    Subcommands: []*cli.Command{
      spot.NewSymbolsCommand(),
      spot.NewTickersCommand(),
      spot.NewDepthCommand(),
      spot.NewWebsocketCommand(),
      spot.NewKlinesCommand(),
      spot.NewIndicatorsCommand(),
      spot.NewStrategiesCommand(),
      spot.NewAccountCommand(),
      spot.NewOrdersCommand(),
      spot.NewPositionsCommand(),
      spot.NewAnalysisCommand(),
      spot.NewPlansCommand(),
      spot.NewLaunchpadCommand(),
      spot.NewScalpingCommand(),
      spot.NewTradingsCommand(),
      spot.NewStreamCommand(),
      spot.NewTasksCommand(),
      spot.NewGamblingCommand(),
    },
  }
}
