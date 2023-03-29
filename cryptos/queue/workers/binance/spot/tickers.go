package spot

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"log"
)

type Tickers struct{}

func NewTickers() *Tickers {
	return &Tickers{}
}

type TickersFlushPayload struct {
	Symbols []string
}

func (h *Tickers) Flush(_ context.Context, t *asynq.Task) error {
	log.Println("Tickers flush...", t.Payload())
	var payload TickersFlushPayload
	json.Unmarshal(t.Payload(), &payload)
	log.Println("symbols", payload.Symbols)
	return nil
}
