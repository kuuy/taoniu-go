package isolated

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/analysis/margin/isolated/tradings"
)

func NewTradingsCommand() *cli.Command {
	return &cli.Command{
		Name:  "tradings",
		Usage: "",
		Subcommands: []*cli.Command{
			tradings.NewFishersCommand(),
		},
	}
}
