package tor

import (
  "context"
  "errors"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "taoniu.local/security/common"
  repositories "taoniu.local/security/repositories/tor"
)

type BridgeHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.BridgesRepository
}

func NewBridgesCommand() *cli.Command {
  var h BridgeHandler
  return &cli.Command{
    Name:  "bridges",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = BridgeHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.BridgesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "import",
        Usage: "",
        Action: func(c *cli.Context) error {
          if c.Args().Get(0) == "" {
            return errors.New("file is empty")
          }
          file := c.Args().Get(0)
          if err := h.Import(file); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "checker",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Checker(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "show",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Show(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *BridgeHandler) Import(file string) error {
  log.Println("tor bridge import...")
  return h.Repository.Import(file)
}

func (h *BridgeHandler) Flush() error {
  log.Println("tor bridge flush...")
  return h.Repository.Flush()
}

func (h *BridgeHandler) Show() error {
  log.Println("tor bridge flush...")
  entity, err := h.Repository.Show()
  if err != nil {
    return nil
  }
  bridge := fmt.Sprintf(
    "%s %s:%d %s cert=%s iat-mode=%d",
    entity.Protocol,
    h.Repository.LongToIp(entity.Ip),
    entity.Port,
    entity.Secret,
    entity.Cert,
    entity.Mode,
  )
  log.Println("bridge", bridge)
  return nil
}

func (h *BridgeHandler) Checker() error {
  log.Println("tor bridge checker...")
  return h.Repository.Checker()
}
