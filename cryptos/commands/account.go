package commands

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/account"
)

func NewAccountCommand() *cli.Command {
	return &cli.Command{
		Name:  "account",
		Usage: "",
		Subcommands: []*cli.Command{
			account.NewUsersCommand(),
		},
	}
}
