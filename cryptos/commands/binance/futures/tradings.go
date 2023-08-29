package futures

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/tradings"
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
