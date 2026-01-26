package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/raydium"
)

func NewRaydiumCommand() *cli.Command {
  return &cli.Command{
    Name:  "raydium",
    Usage: "",
    Subcommands: []*cli.Command{
      raydium.NewSwapCommand(),
    },
  }
}
