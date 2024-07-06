package margin

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/margin/isolated"
)

func NewIsoLatedCommand() *cli.Command {
  return &cli.Command{
    Name:  "isolated",
    Usage: "",
    Subcommands: []*cli.Command{
      isolated.NewAccountCommand(),
      isolated.NewOrdersCommand(),
      isolated.NewTradingsCommand(),
      isolated.NewWebsocketCommand(),
    },
  }
}
