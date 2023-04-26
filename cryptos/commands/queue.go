package commands

import (
	"log"
	"taoniu.local/cryptos/common"

	"github.com/hibiken/asynq"
	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/queue"
)

type QueueHandler struct{}

func NewQueueCommand() *cli.Command {
	var h QueueHandler
	return &cli.Command{
		Name:  "queue",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = QueueHandler{}
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

func (h *QueueHandler) run() error {
	log.Println("queue running...")

	worker := common.NewAsynqServer()

	mux := asynq.NewServeMux()
	queue.NewWorkers().Register(mux)
	if err := worker.Run(mux); err != nil {
		return err
	}

	return nil
}
