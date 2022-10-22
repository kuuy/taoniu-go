package spot

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/plans"
)

func NewPlansCommand() *cli.Command {
	return &cli.Command{
		Name:  "plans",
		Usage: "",
		Subcommands: []*cli.Command{
			plans.NewDailyCommand(),
		},
	}
}
