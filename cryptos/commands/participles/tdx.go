package participles

import (
	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/commands/participles/tdx"
)

func NewTdxCommand() *cli.Command {
	return tdx.NewTdxCommand()
}
