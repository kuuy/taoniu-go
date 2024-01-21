package commands

import (
  "crypto/md5"
  "encoding/hex"
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/account/common"
  "taoniu.local/account/repositories"
)

type UsersHandler struct {
  Db         *gorm.DB
  Repository *repositories.UsersRepository
}

func NewUsersCommand() *cli.Command {
  var h UsersHandler
  return &cli.Command{
    Name:  "users",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = UsersHandler{
        Db: common.NewDB(),
      }
      h.Repository = &repositories.UsersRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "create",
        Usage: "",
        Action: func(c *cli.Context) error {
          email := c.Args().Get(0)
          if email == "" {
            log.Fatal("email can not be empty")
            return nil
          }
          password := c.Args().Get(1)
          if password == "" {
            log.Fatal("password can not be empty")
            return nil
          }
          if err := h.Create(email, password); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *UsersHandler) Create(email string, password string) error {
  log.Println("users create...")

  hash := md5.Sum([]byte(password))
  password = hex.EncodeToString(hash[:])

  return h.Repository.Create(email, password)
}
