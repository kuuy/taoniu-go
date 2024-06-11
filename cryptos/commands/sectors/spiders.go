package sectors

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/sectors/spiders"
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
