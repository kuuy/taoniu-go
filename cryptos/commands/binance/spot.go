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
      spot.NewMarginCommand(),
      spot.NewAnalysisCommand(),
      spot.NewPlansCommand(),
      spot.NewGridsCommand(),
      spot.NewFishersCommand(),
      spot.NewTradingsCommand(),
      spot.NewStreamCommand(),
    },
  }
}
