package tor

import (
  "context"
  "errors"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "strconv"
  "taoniu.local/security/common"
  repositories "taoniu.local/security/repositories/tor"
)

type ProxiesHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ProxiesRepository
}

func NewProxiesCommand() *cli.Command {
  var h ProxiesHandler
  return &cli.Command{
    Name:  "proxies",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ProxiesHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.ProxiesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if c.Args().Get(0) == "" {
        return errors.New("id is empty")
      }
      id, err := strconv.Atoi(c.Args().Get(0))
      if err != nil {
        return err
      }
      if err := h.start(id); err != nil {
        return cli.Exit(err.Error(), 1)
      }

      return nil
    },
  }
}

func (h *ProxiesHandler) start(id int) error {
  h.Repository.Start(id)
  return nil
}
