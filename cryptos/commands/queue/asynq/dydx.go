package asynq

import (
  "log"

  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers"
)

type DydxHandler struct{}

func NewDydxCommand() *cli.Command {
  var h DydxHandler
  return &cli.Command{
    Name:  "dydx",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DydxHandler{}
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *DydxHandler) run() error {
  log.Println("queue running...")

  worker := common.NewAsynqServer("DYDX")

  mux := asynq.NewServeMux()
  workers.NewDydx().Register(mux)
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
