package spot

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/klines"
)

func NewKlinesCommand() *cli.Command {
	return &cli.Command{
		Name:  "klines",
		Usage: "",
		Subcommands: []*cli.Command{
			klines.NewDailyCommand(),
		},
	}
}
