package raydium

import (
	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/commands/raydium/perpetuals"
)

func NewPerpetualsCommand() *cli.Command {
	return &cli.Command{
		Name:  "perpetuals",
		Usage: "",
		Subcommands: []*cli.Command{
			perpetuals.NewKlinesCommand(),
		},
	}
}
