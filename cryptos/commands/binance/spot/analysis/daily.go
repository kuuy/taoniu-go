package analysis

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/analysis/daily"
)

func NewDailyCommand() *cli.Command {
	return &cli.Command{
		Name:  "daily",
		Usage: "",
		Subcommands: []*cli.Command{
			daily.NewMarginCommand(),
		},
	}
}
