package wolf

import (
	"github.com/urfave/cli/v2"
	"log"
	repositories "taoniu.local/gamblings/repositories/wolf"
)

type AccountHandler struct {
	Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
	var h AccountHandler
	return &cli.Command{
		Name:  "account",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = AccountHandler{}
			h.Repository = &repositories.AccountRepository{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "balance",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.balance(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *AccountHandler) balance() error {
	log.Println("wolf account balance...")
	err := h.Repository.Balance()
	if err != nil {
		log.Println("wolf account balance error", err)
	}
	return nil
}
