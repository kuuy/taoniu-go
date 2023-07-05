package queue

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/queue/nats"
)

func NewNatsCommand() *cli.Command {
  return &cli.Command{
    Name:  "nats",
    Usage: "",
    Subcommands: []*cli.Command{
      nats.NewBinanceCommand(),
    },
  }
}
