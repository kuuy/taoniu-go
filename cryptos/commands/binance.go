package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance"
)

func NewBinanceCommand() *cli.Command {
	return &cli.Command{
		Name:  "binance",
		Usage: "",
		Subcommands: []*cli.Command{
			binance.NewSymbolsCommand(),
			binance.NewSpotCommand(),
			binance.NewSavingsCommand(),
		},
	}
}
