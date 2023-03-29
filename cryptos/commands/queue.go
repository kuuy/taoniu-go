package commands

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/urfave/cli/v2"
	"log"
	config "taoniu.local/cryptos/config/queue"
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

	rdb := asynq.RedisClientOpt{
		Addr: config.REDIS_ADDR,
		DB:   config.REDIS_DB,
	}
	worker := asynq.NewServer(rdb, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			config.BINANCE_SPOT_TICKERS: 10,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(h.reportError),
	})

	mux := asynq.NewServeMux()
	queue.NewWorkers().Register(mux)
	if err := worker.Run(mux); err != nil {
		return err
	}

	return nil
}

func (h *QueueHandler) reportError(ctx context.Context, task *asynq.Task, err error) {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	if retried >= maxRetry {
		err = fmt.Errorf("retry exhausted for task %s: %w", task.Type, err)
	}
	log.Println("err", err)
}
