package wolf

import (
	"github.com/urfave/cli/v2"
	hilo "taoniu.local/gamblings/commands/wolf/hilo"
)

func NewHiloCommand() *cli.Command {
	return &cli.Command{
		Name:  "hilo",
		Usage: "",
		Subcommands: []*cli.Command{
			hilo.NewBetCommand(),
		},
	}
}
