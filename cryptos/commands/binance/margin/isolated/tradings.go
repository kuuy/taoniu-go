package isolated

import (
  "github.com/urfave/cli/v2"
)

func NewTradingsCommand() *cli.Command {
  return &cli.Command{
    Name:  "tradings",
    Usage: "",
    Subcommands: []*cli.Command{},
  }
}
