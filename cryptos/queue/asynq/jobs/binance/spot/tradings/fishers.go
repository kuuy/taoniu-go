package tradings

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Fishers struct{}

type FishersFlushPayload struct {
	Symbol string
}

type FishersPlacePayload struct {
	Symbol string
}

func (h *Fishers) Flush(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(FishersFlushPayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tradings:fishers:flush", payload), nil
}

func (h *Fishers) Place(symbol string) (*asynq.Task, error) {
	payload, err := json.Marshal(FishersPlacePayload{symbol})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tradings:fishers:place", payload), nil
}
