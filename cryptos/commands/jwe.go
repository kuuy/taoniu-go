package commands

import (
  "errors"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/repositories"
)

type JweHandler struct {
  JweRepository *repositories.JweRepository
}

func NewJweCommand() *cli.Command {
  var h JweHandler
  return &cli.Command{
    Name:  "jwe",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = JweHandler{}
      h.JweRepository = &repositories.JweRepository{}
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "encrypt",
        Usage: "",
        Action: func(c *cli.Context) (err error) {
          payload := c.Args().Get(0)
          if payload == "" {
            err = errors.New("payload can not be empty")
            return
          }
          if err = h.Encrypt(payload); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return
        },
      },
      {
        Name:  "decrypt",
        Usage: "",
        Action: func(c *cli.Context) (err error) {
          jweCompact := c.Args().Get(0)
          if jweCompact == "" {
            err = errors.New("payload can not be empty")
            return
          }
          if err = h.Decrypt(jweCompact); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return
        },
      },
    },
  }
}

func (h *JweHandler) Encrypt(payload string) error {
  log.Println("jwe encrypt...")
  jweCompact, _ := h.JweRepository.Encrypt([]byte(payload))
  log.Println("jweCompact", jweCompact)
  return nil
}

func (h *JweHandler) Decrypt(jweCompact string) (err error) {
  log.Println("jwe decrypt...")
  payload, err := h.JweRepository.Decrypt(jweCompact)
  if err != nil {
    return
  }
  log.Println("payload", string(payload))
  return
}
