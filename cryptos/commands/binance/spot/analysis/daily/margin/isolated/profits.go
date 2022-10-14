package isolated

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/analysis/daily/margin/isolated/profits"
)

func NewProfitsCommand() *cli.Command {
	return &cli.Command{
		Name:  "profits",
		Usage: "",
		Subcommands: []*cli.Command{
			profits.NewDailyCommand(),
		},
	}
}
