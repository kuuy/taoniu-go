package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/security/commands/gfw"
)

func NewGfwCommand() *cli.Command {
	return &cli.Command{
		Name:  "gfw",
		Usage: "",
		Subcommands: []*cli.Command{
			gfw.NewCrawlCommand(),
			gfw.NewDnsCommand(),
		},
	}
}
