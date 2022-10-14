package daily

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/analysis/daily/margin"
)

func NewMarginCommand() *cli.Command {
	return &cli.Command{
		Name:  "margin",
		Usage: "",
		Subcommands: []*cli.Command{
			margin.NewIsolatedCommand(),
		},
	}
}
