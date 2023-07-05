package binance

import (
  "log"

  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  workers "taoniu.local/cryptos/queue/asynq/workers/binance"
)

type FuturesHandler struct{}

func NewFuturesCommand() *cli.Command {
  var h FuturesHandler
  return &cli.Command{
    Name:  "futures",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = FuturesHandler{}
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

func (h *FuturesHandler) run() error {
  log.Println("queue running...")

  worker := common.NewAsynqServer()

  mux := asynq.NewServeMux()
  workers.NewFutures().Register(mux)
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
