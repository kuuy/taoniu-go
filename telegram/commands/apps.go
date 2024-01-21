package commands

import (
  "context"
  "errors"
  "log"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/telegram/common"
  "taoniu.local/telegram/repositories"
)

type AppsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AppsRepository
}

func NewAppsCommand() *cli.Command {
  var h AppsHandler
  return &cli.Command{
    Name:  "apps",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AppsHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AppsRepository{
        Db:  h.Db,
        Ctx: h.Ctx,
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      phone := c.Args().Get(0)
      if phone == "" {
        return errors.New("phone is empty")
      }
      appID, _ := strconv.Atoi(c.Args().Get(1))
      if appID == 0 {
        return errors.New("appID is empty")
      }
      appHash := c.Args().Get(2)
      if appHash == "" {
        return errors.New("appHash is empty")
      }
      if err := h.Run(phone, appID, appHash); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *AppsHandler) Run(phone string, appID int, appHash string) error {
  log.Println("apps run...")
  return h.Repository.Run(phone, appID, appHash)
}
