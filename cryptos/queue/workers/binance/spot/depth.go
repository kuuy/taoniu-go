package spot

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Depth struct{}

func NewDepth() *Depth {
	return &Depth{}
}

type DepthFlushPayload struct {
	Symbol string
}

func (h *Depth) Flush(_ context.Context, t *asynq.Task) error {
	var payload DepthFlushPayload
	json.Unmarshal(t.Payload(), &payload)
	repository := &repositories.DepthRepository{
		Db:       common.NewDB(),
		UseProxy: true,
	}
	repository.Flush(payload.Symbol)

	return nil
}
