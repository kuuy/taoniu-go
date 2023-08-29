package binance

import (
  "log"

  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  workers "taoniu.local/cryptos/queue/asynq/workers/binance"
)

type SpotHandler struct{}

func NewSpotCommand() *cli.Command {
  var h SpotHandler
  return &cli.Command{
    Name:  "spot",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SpotHandler{}
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

func (h *SpotHandler) run() error {
  log.Println("queue running...")

  worker := common.NewAsynqServer("BINANCE_SPOT")

  mux := asynq.NewServeMux()
  workers.NewSpot().Register(mux)
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
