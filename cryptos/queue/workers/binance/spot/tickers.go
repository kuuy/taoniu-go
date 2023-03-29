package spot

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Tickers struct{}

func NewTickers() *Tickers {
	return &Tickers{}
}

type TickersFlushPayload struct {
	Symbols []string
}

func (h *Tickers) Flush(ctx context.Context, t *asynq.Task) error {
	var payload TickersFlushPayload
	json.Unmarshal(t.Payload(), &payload)
	repository := &repositories.TickersRepository{
		Rdb:      common.NewRedis(),
		Ctx:      ctx,
		UseProxy: true,
	}
	repository.Flush(payload.Symbols)

	return nil
}
