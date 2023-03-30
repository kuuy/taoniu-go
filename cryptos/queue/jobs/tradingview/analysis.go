package tradingview

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type Analysis struct{}

type AnalysisFlushPayload struct {
	Exchange string
	Symbol   string
	Interval string
	UseProxy bool
}

func (h *Analysis) Flush(exchange string, symbol string, interval string, useProxy bool) (*asynq.Task, error) {
	payload, err := json.Marshal(AnalysisFlushPayload{exchange, symbol, interval, useProxy})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("tradingview:analysis:flush", payload), nil
}
