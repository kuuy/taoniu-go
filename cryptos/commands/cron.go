package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/cron"
)

func NewCronCommand() *cli.Command {
  return &cli.Command{
    Name:  "cron",
    Usage: "",
    Subcommands: []*cli.Command{
      cron.NewBinanceCommand(),
    },
  }
}
