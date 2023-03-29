package spot

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Tickers struct{}

type TickersFlushPayload struct {
	Symbols []string
}

func (h *Tickers) Flush(symbols []string) (*asynq.Task, error) {
	payload, err := json.Marshal(TickersFlushPayload{symbols})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:tickers:flush", payload), nil
}
