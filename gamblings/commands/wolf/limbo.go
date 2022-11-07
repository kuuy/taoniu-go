package wolf

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/gamblings/commands/wolf/limbo"
)

func NewLimboCommand() *cli.Command {
	return &cli.Command{
		Name:  "limbo",
		Usage: "",
		Subcommands: []*cli.Command{
			limbo.NewBetCommand(),
		},
	}
}
