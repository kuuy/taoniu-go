package gfw

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	"taoniu.local/security/common"
	repositories "taoniu.local/security/repositories/gfw"
)

type DnsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DnsRepository
}

func NewDnsCommand() *cli.Command {
	var h DnsHandler
	return &cli.Command{
		Name:  "dns",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = DnsHandler{}
			h.Repository = &repositories.DnsRepository{
				Db: common.NewDB(),
			}
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
			{
				Name:  "cache",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.cache(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "submit",
				Usage: "",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						return errors.New("domains is empty")
					}
					domains := strings.Split(c.Args().Get(0), ",")
					if err := h.submit(domains); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "monitor",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.monitor(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DnsHandler) flush() error {
	log.Println("gfw dns flush...")
	return h.Repository.Flush()
}

func (h *DnsHandler) cache() error {
	log.Println("gfw dns cache...")
	return h.Repository.Cache()
}

func (h *DnsHandler) submit(domains []string) error {
	log.Println("gfw dns submit...")
	return h.Repository.Submit(domains)
}

func (h *DnsHandler) monitor() error {
	log.Println("gfw dns monitor...")
	return h.Repository.Monitor()
}
