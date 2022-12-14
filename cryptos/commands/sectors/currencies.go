package sectors

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/sectors/currencies/spiders"
)

func NewCurrenciesCommand() *cli.Command {
	return &cli.Command{
		Name:  "currencies",
		Usage: "",
		Subcommands: []*cli.Command{
			spiders.NewSourcesCommand(),
			spiders.NewCrawlsCommand(),
		},
	}
}
