package spot

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/indicators"
)

func NewIndicatorsCommand() *cli.Command {
	return &cli.Command{
		Name:  "indicators",
		Usage: "",
		Subcommands: []*cli.Command{
			indicators.NewDailyCommand(),
		},
	}
}
