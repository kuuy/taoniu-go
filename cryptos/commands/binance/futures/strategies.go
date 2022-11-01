package futures

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/futures/strategies"
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
