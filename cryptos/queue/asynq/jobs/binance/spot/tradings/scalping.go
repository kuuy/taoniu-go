package tradings

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Scalping struct{}

type ScalpingFlushPayload struct {
	Symbol string
}

type ScalpingPlacePayload struct {
	Symbol string
}

func (h *Scalping) Flush(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(ScalpingFlushPayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tradings:scalping:flush", payload), nil
}

func (h *Scalping) Place(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(ScalpingPlacePayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tradings:scalping:place", payload), nil
}
