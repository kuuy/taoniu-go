package analysis

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/analysis/tradings"
)

func NewTradingsCommand() *cli.Command {
  return &cli.Command{
    Name:  "tradings",
    Usage: "",
    Subcommands: []*cli.Command{
      tradings.NewTriggersCommand(),
      tradings.NewScalpingCommand(),
    },
  }
}
