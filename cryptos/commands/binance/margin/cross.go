package margin

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/commands/binance/margin/cross"
)

func NewCrossCommand() *cli.Command {
  return &cli.Command{
    Name:  "cross",
    Usage: "",
    Subcommands: []*cli.Command{
      cross.NewAccountCommand(),
      cross.NewOrdersCommand(),
      cross.NewScalpingCommand(),
      cross.NewTradingsCommand(),
      cross.NewAnalysisCommand(),
      cross.NewTasksCommand(),
    },
  }
}
