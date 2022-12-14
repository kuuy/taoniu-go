package exchanges

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/exchanges/spiders"
)

func NewSpidersCommand() *cli.Command {
	return &cli.Command{
		Name:  "spiders",
		Usage: "",
		Subcommands: []*cli.Command{
			spiders.NewSourcesCommand(),
			spiders.NewCrawlsCommand(),
		},
	}
}
