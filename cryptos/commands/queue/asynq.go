package queue

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/queue/asynq"
)

func NewAsynqCommand() *cli.Command {
  return &cli.Command{
    Name:  "asynq",
    Usage: "",
    Subcommands: []*cli.Command{
      asynq.NewBinanceCommand(),
    },
  }
}
