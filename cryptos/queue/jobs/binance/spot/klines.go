package spot

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Klines struct{}

type KlinesFlushPayload struct {
	Symbol   string
	Interval string
	Limit    int
	UseProxy bool
}

func (h *Klines) Flush(symbol string, interval string, limit int, useProxy bool) (*asynq.Task, error) {
	payload, err := json.Marshal(KlinesFlushPayload{symbol, interval, limit, useProxy})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("binance:spot:klines:flush", payload), nil
}
