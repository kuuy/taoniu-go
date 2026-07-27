package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/grpc"
)

func NewGrpcCommand() *cli.Command {
	return &cli.Command{
		Name:  "grpc",
		Usage: "",
		Subcommands: []*cli.Command{
			grpc.NewBinanceCommand(),
		},
	}
}
