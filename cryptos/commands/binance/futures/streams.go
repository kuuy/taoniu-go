package futures

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/futures/streams"
)

type StreamsHandler struct{}

func NewStreamCommand() *cli.Command {
  return &cli.Command{
    Name:  "streams",
    Usage: "",
    Subcommands: []*cli.Command{
      streams.NewAccountCommand(),
    },
  }
}
