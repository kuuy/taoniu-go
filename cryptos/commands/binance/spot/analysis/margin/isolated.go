package margin

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/analysis/margin/isolated"
)

func NewIsolatedCommand() *cli.Command {
	return &cli.Command{
		Name:  "isolated",
		Usage: "",
		Subcommands: []*cli.Command{
			isolated.NewProfitsCommand(),
		},
	}
}
