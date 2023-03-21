package tradings

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/binance/spot/tradings/fishers"
)

func NewFishersCommand() *cli.Command {
	command := fishers.NewFishersCommand()
	command.Subcommands = append(
		command.Subcommands,
		fishers.NewGridsCommand(),
	)
	return command
}
