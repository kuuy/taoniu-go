package grpc

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/grpc/binance"
)

func NewBinanceCommand() *cli.Command {
	return &cli.Command{
		Name:  "binance",
		Usage: "",
		Subcommands: []*cli.Command{
			binance.NewSpotCommand(),
			binance.NewFuturesCommand(),
		},
	}
}
