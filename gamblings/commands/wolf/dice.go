package wolf

import (
	"github.com/urfave/cli/v2"
	dice "taoniu.local/gamblings/commands/wolf/dice"
)

func NewDiceCommand() *cli.Command {
	return &cli.Command{
		Name:  "dice",
		Usage: "",
		Subcommands: []*cli.Command{
			dice.NewHuntCommand(),
			dice.NewBetCommand(),
			dice.NewMultipleCommand(),
			dice.NewPlansCommand(),
			dice.NewHellsCommand(),
		},
	}
}
