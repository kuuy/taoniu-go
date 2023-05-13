package commands

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/telegram/common"
  "taoniu.local/telegram/repositories"
)

type BotsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.BotsRepository
}

func NewBotsCommand() *cli.Command {
  var h BotsHandler
  return &cli.Command{
    Name:  "bots",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = BotsHandler{
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.BotsRepository{}
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "auth",
        Usage: "",
        Action: func(c *cli.Context) error {
          token := c.Args().Get(0)
          if token == "" {
            log.Fatal("token can not be empty")
            return nil
          }
          if err := h.Auth(token); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *BotsHandler) Auth(token string) error {
  log.Println("bots auth...")

  return h.Repository.Auth(token)
}
