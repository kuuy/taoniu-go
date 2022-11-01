package binance

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/futures"
)

func NewFutoresCommand() *cli.Command {
	return &cli.Command{
		Name:  "futures",
		Usage: "",
		Subcommands: []*cli.Command{
			futures.NewSymbolsCommand(),
			futures.NewTickersCommand(),
			futures.NewStreamCommand(),
			futures.NewWebsocketCommand(),
			futures.NewKlinesCommand(),
			futures.NewIndicatorsCommand(),
			futures.NewStrategiesCommand(),
			futures.NewAccountCommand(),
			futures.NewOrdersCommand(),
			futures.NewPlansCommand(),
			futures.NewGridsCommand(),
			futures.NewTradingsCommand(),
		},
	}
}
