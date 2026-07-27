package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/api"
)

func NewApiCommand() *cli.Command {
	return &cli.Command{
		Name:  "api",
		Usage: "",
		Subcommands: []*cli.Command{
			api.NewBinanceCommand(),
		},
	}
}
