package futures

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/indicators"
)

func NewIndicatorsCommand() *cli.Command {
  return &cli.Command{
    Name:  "indicators",
    Usage: "",
    Subcommands: []*cli.Command{
      indicators.NewDailyCommand(),
      indicators.NewMinutelyCommand(),
    },
  }
}
