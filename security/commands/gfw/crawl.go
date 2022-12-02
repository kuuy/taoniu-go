package gfw

import (
	"log"

	"github.com/urfave/cli/v2"

	repositories "taoniu.local/security/repositories/gfw"
)

type CrawlHandler struct {
	Repository *repositories.CrawlRepository
}

func NewCrawlCommand() *cli.Command {
	var h CrawlHandler
	return &cli.Command{
		Name:  "crawl",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = CrawlHandler{}
			h.Repository = &repositories.CrawlRepository{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *CrawlHandler) flush() error {
	log.Println("gfw crawl flush...")
	ab, err := h.Repository.Flush()
	if err != nil {
		return err
	}

	isBlock := ab.ShouldBlock("l.google.com.", nil)
	log.Println("block", isBlock)

	return nil
}
