package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/sectors"
)

func NewSectorsCommand() *cli.Command {
	return &cli.Command{
		Name:  "sectors",
		Usage: "",
		Subcommands: []*cli.Command{
			sectors.NewSpidersCommand(),
		},
	}
}
