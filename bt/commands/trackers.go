package commands

import (
	"github.com/urfave/cli/v2"
	"log"
	"strings"
	"taoniu.local/bt/repositories"
)

type TrackersHandler struct {
	Repository *repositories.TrackersRepository
}

func NewTrackersCommands() *cli.Command {
	var h TrackersHandler
	return &cli.Command{
		Name:  "trackers",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = TrackersHandler{}
			h.Repository = &repositories.TrackersRepository{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "black",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Black(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "crawl",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Crawl(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *TrackersHandler) Black() error {
	trackers, err := h.Repository.Black()
	if err != nil {
		return err
	}
	log.Println("black trackers", strings.Join(trackers, ","))
	return nil
}

func (h *TrackersHandler) Crawl() error {
	trackers, err := h.Repository.Crawl()
	if err != nil {
		return err
	}
	log.Println("trackers", strings.Join(trackers, ","))
	return nil
}
