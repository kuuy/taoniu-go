package queue

import (
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers"
)

type Workers struct{}

func NewWorkers() *Workers {
	return &Workers{}
}

func (h *Workers) Register(mux *asynq.ServeMux) error {
	workers.NewBinance().Register(mux)
	return nil
}
