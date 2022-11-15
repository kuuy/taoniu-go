package commands

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/account/common"
	repositories "taoniu.local/account/repositories"
)

type UsersHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.UsersRepository
}

func NewUsersCommand() *cli.Command {
	var h UsersHandler
	return &cli.Command{
		Name:  "users",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = UsersHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.UsersRepository{
				Db: pool.NewDB(),
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "",
				Action: func(c *cli.Context) error {
					email := c.Args().Get(0)
					password := c.Args().Get(1)
					if email == "" {
						log.Fatal("email can not be empty")
						return nil
					}
					if password == "" {
						log.Fatal("password can not be empty")
						return nil
					}
					if err := h.create(email, password); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *UsersHandler) create(email string, password string) error {
	log.Println("users create...")

	hash := md5.Sum([]byte(password))
	password = hex.EncodeToString(hash[:])

	return h.Repository.Create(email, password)
}
