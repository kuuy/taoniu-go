package tradings

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/analysis/margin/isolated/tradings/fishers"
)

func NewFishersCommand() *cli.Command {
	return &cli.Command{
		Name:  "fishers",
		Usage: "",
		Subcommands: []*cli.Command{
			fishers.NewGridsCommand(),
		},
	}
}
