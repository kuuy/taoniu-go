package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/exchanges"
)

func NewExchangesCommand() *cli.Command {
	return &cli.Command{
		Name:  "exchanges",
		Usage: "",
		Subcommands: []*cli.Command{
			exchanges.NewSpidersCommand(),
		},
	}
}
