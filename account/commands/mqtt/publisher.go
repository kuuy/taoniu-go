package mqtt

import (
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/account/repositories"
)

type PublishersHandler struct {
  TokenRepository *repositories.TokenRepository
}

func NewPublishersCommand() *cli.Command {
  var h PublishersHandler
  return &cli.Command{
    Name:  "publishers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PublishersHandler{}
      h.TokenRepository = &repositories.TokenRepository{}
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "access-token",
        Usage: "",
        Action: func(c *cli.Context) error {
          id := c.Args().Get(0)
          if id == "" {
            log.Fatal("id can not be empty")
            return nil
          }
          if err := h.Token(id); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *PublishersHandler) Token(id string) (err error) {
  log.Println("publishers access token...")
  accessToken, err := h.TokenRepository.AccessToken(id)
  log.Fatalln("publishers access token", accessToken)
  return
}
