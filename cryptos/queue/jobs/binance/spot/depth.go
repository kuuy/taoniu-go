package spot

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Depth struct{}

type DepthFlushPayload struct {
	Symbol string
}

func (h *Depth) Flush(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(DepthFlushPayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:depth:flush", payload), nil
}
