package commands

import (
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
		Concurrency: 30,
		Queues: map[string]int{
			config.BINANCE_SPOT_DEPTH:         5,
			config.BINANCE_SPOT_DEPTH_DELAY:   5,
			config.BINANCE_SPOT_TICKERS:       3,
			config.BINANCE_SPOT_TICKERS_DELAY: 3,
			config.BINANCE_SPOT_KLINES:        5,
			config.BINANCE_SPOT_KLINES_DELAY:  5,
			config.TRADINGVIEW_ANALYSIS:       8,
			config.TRADINGVIEW_ANALYSIS_DELAY: 8,
		},
	})

	mux := asynq.NewServeMux()
	queue.NewWorkers().Register(mux)
	if err := worker.Run(mux); err != nil {
		return err
	}

	return nil
}
