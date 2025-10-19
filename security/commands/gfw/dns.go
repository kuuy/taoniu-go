package gfw

import (
  "context"
  "errors"
  "log"
  "strings"

  "github.com/urfave/cli/v2"

  "taoniu.local/security/common"
  "taoniu.local/security/grpc/services"
  repositories "taoniu.local/security/repositories/gfw"
)

type DnsHandler struct {
  Ctx        context.Context
  Repository *repositories.DnsRepository
}

func NewDnsCommand() *cli.Command {
  var h DnsHandler
  return &cli.Command{
    Name:  "dns",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DnsHandler{
        Ctx: context.Background(),
      }
      h.Repository = &repositories.DnsRepository{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: h.Ctx,
      }
      h.Repository.Service = &services.Aes{
        Ctx: h.Ctx,
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
        Name:  "lookup",
        Usage: "",
        Action: func(c *cli.Context) error {
          if c.Args().Get(0) == "" {
            return errors.New("domain is empty")
          }
          domain := strings.TrimSpace(c.Args().Get(0))
          if err := h.lookup(domain); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "query",
        Usage: "",
        Action: func(c *cli.Context) error {
          if c.Args().Get(0) == "" {
            return errors.New("domains is empty")
          }
          domains := strings.Split(c.Args().Get(0), ",")
          if err := h.query(domains); err != nil {
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

func (h *DnsHandler) lookup(domain string) (err error) {
  log.Println("gfw dns lookup...")
  var result []string
  result, err = h.Repository.Lookup(domain)
  if err != nil {
    log.Println("dns lookup failed", domain, err)
    return
  }
  log.Println("result", result)
  return
}

func (h *DnsHandler) query(domains []string) (err error) {
  log.Println("gfw dns query...")
  var result []string
  result, err = h.Repository.Query(domains)
  if err != nil {
    return
  }
  log.Println("result", result)
  return
}

func (h *DnsHandler) submit(domains []string) error {
  log.Println("gfw dns submit...")
  return h.Repository.Submit(domains)
}

func (h *DnsHandler) monitor() error {
  log.Println("gfw dns monitor...")
  return h.Repository.Monitor()
}
