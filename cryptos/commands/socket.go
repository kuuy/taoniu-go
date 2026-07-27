package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/socket"
)

func NewSocketCommand() *cli.Command {
	return &cli.Command{
		Name:  "socket",
		Usage: "",
		Subcommands: []*cli.Command{
			socket.NewBinanceCommand(),
		},
	}
}
