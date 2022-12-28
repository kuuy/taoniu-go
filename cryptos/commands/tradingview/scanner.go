package tradingview

import (
	"github.com/urfave/cli/v2"
	"log"
	repositories "taoniu.local/cryptos/repositories/tradingview"
)

type ScannerHandler struct {
	Repository *repositories.ScannerRepository
}

func NewScannerCommand() *cli.Command {
	var h ScannerHandler
	return &cli.Command{
		Name:  "scanner",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = ScannerHandler{}
			h.Repository = &repositories.ScannerRepository{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "scan",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Scan(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *ScannerHandler) Scan() error {
	log.Println("scanner scan processing...")
	summary, err := h.Repository.Scan("BINANCE", "AVAXBUSD", "1m")
	if err != nil {
		return err
	}
	log.Println("scan", summary)
	return nil
}
