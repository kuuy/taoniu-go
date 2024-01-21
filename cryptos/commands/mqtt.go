package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/mqtt"
)

func NewMqttCommand() *cli.Command {
  return &cli.Command{
    Name:  "mqtt",
    Usage: "",
    Subcommands: []*cli.Command{
      mqtt.NewBinanceCommand(),
    },
  }
}
