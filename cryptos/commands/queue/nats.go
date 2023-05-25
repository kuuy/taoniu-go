package queue

import (
  "log"
  "sync"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  queue "taoniu.local/cryptos/queue/nats"
)

type NatsHandler struct{}

func NewNatsCommand() *cli.Command {
  var h NatsHandler
  return &cli.Command{
    Name:  "nats",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = NatsHandler{}
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

func (h *NatsHandler) run() error {
  log.Println("nats running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  nc := common.NewNats()
  defer nc.Close()

  queue.NewWorkers().Subscribe(nc)

  <-h.wait(wg)

  return nil
}

func (h *NatsHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
