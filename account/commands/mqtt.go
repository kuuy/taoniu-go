package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/account/commands/mqtt"
)

func NewMqttCommand() *cli.Command {
  return &cli.Command{
    Name:  "mqtt",
    Usage: "",
    Subcommands: []*cli.Command{
      mqtt.NewPublishersCommand(),
    },
  }
}
