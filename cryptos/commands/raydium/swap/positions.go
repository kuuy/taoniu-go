package swap

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories/raydium/swap"
)

func NewPositionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "positions",
		Usage: "Raydium swap positions management",
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "Flush positions from transactions",
				Action: func(c *cli.Context) error {
					db := common.NewDB(3)
					mintsRepo := &swap.MintsRepository{
						Db:  db,
						Ctx: c.Context,
					}
					repo := &swap.PositionsRepository{
						Db:              db,
						Ctx:             c.Context,
						MintsRepository: mintsRepo,
					}
					if err := repo.Flush(); err != nil {
						return err
					}
					fmt.Println("Positions flushed successfully")
					return nil
				},
			},
		},
	}
}
