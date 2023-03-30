package spot

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Klines struct {
	Db         *gorm.DB
	Repository *repositories.KlinesRepository
}

func NewKlines() *Klines {
	h := &Klines{
		Db: common.NewDB(),
	}
	h.Repository = &repositories.KlinesRepository{
		Db: h.Db,
	}
	return h
}

type KlinesFlushPayload struct {
	Symbol   string
	Interval string
	Limit    int
	UseProxy bool
}

func (h *Klines) Flush(ctx context.Context, t *asynq.Task) error {
	var payload KlinesFlushPayload
	json.Unmarshal(t.Payload(), &payload)

	if payload.UseProxy {
		h.Repository.UseProxy = true
	}

	h.Repository.Flush(payload.Symbol, payload.Interval, payload.Limit)

	return nil
}
