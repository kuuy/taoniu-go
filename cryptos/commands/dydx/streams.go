package dydx

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/dydx/streams"
)

type StreamsHandler struct{}

func NewStreamCommand() *cli.Command {
  return &cli.Command{
    Name:  "streams",
    Usage: "",
    Subcommands: []*cli.Command{
      streams.NewTradesCommand(),
      streams.NewAccountCommand(),
    },
  }
}
