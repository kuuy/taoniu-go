package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/participles"
)

func NewParticiplesCommand() *cli.Command {
	return &cli.Command{
		Name:  "participles",
		Usage: "",
		Subcommands: []*cli.Command{
			participles.NewBasicCommand(),
			participles.NewTdxCommand(),
		},
	}
}
