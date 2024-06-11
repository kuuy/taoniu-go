package participles

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/participles/basic"
)

func NewBasicCommand() *cli.Command {
  return basic.NewBasicCommand()
}
