package currencies

import (
	"github.com/urfave/cli/v2"
	spiders2 "taoniu.local/cryptos/commands/binance/currencies/spiders"
)

func NewSpidersCommand() *cli.Command {
	return &cli.Command{
		Name:  "spiders",
		Usage: "",
		Subcommands: []*cli.Command{
			spiders2.NewSourcesCommand(),
			spiders2.NewCrawlsCommand(),
		},
	}
}
