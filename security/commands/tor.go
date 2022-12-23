package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/security/commands/tor"
)

func NewTorCommand() *cli.Command {
	return &cli.Command{
		Name:  "tor",
		Usage: "",
		Subcommands: []*cli.Command{
			tor.NewBridgesCommand(),
			tor.NewProxiesCommand(),
		},
	}
}
