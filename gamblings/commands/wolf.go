package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/gamblings/commands/wolf"
)

func NewWolfCommand() *cli.Command {
	return &cli.Command{
		Name:  "wolf",
		Usage: "",
		Subcommands: []*cli.Command{
			wolf.NewAccountCommand(),
			wolf.NewDiceCommand(),
			wolf.NewHiloCommand(),
			wolf.NewLimboCommand(),
		},
	}
}
