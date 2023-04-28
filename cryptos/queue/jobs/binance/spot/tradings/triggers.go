package tradings

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Triggers struct{}

type TriggersFlushPayload struct {
	Symbol string
}

type TriggersPlacePayload struct {
	Symbol string
}

func (h *Triggers) Flush(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(TriggersFlushPayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tradings:triggers:flush", payload), nil
}

func (h *Triggers) Place(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(TriggersPlacePayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tradings:triggers:place", payload), nil
}
