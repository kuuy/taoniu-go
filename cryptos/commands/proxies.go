package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/proxies"
)

func NewProxiesCommand() *cli.Command {
	return &cli.Command{
		Name:  "proxies",
		Usage: "",
		Subcommands: []*cli.Command{
			proxies.NewCrawlsCommand(),
		},
	}
}
