package spot

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/spot/streams"
)

type StreamsHandler struct{}

func NewStreamCommand() *cli.Command {
  return &cli.Command{
    Name:  "streams",
    Usage: "",
    Subcommands: []*cli.Command{
      streams.NewAccountCommand(),
      streams.NewTickersCommand(),
    },
  }
}
