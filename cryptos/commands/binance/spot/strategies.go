package spot

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/strategies"
)

func NewStrategiesCommand() *cli.Command {
	return &cli.Command{
		Name:  "strategies",
		Usage: "",
		Subcommands: []*cli.Command{
			strategies.NewDailyCommand(),
		},
	}
}
