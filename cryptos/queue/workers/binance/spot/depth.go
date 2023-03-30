package spot

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Depth struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	Repository        *repositories.DepthRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func NewDepth() *Depth {
	h := &Depth{
		Db:  common.NewDB(),
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.DepthRepository{
		Db: h.Db,
	}
	h.SymbolsRepository = &repositories.SymbolsRepository{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}
	return h
}

type DepthFlushPayload struct {
	Symbol   string
	UseProxy bool
}

func (h *Depth) Flush(ctx context.Context, t *asynq.Task) error {
	var payload DepthFlushPayload
	json.Unmarshal(t.Payload(), &payload)

	if payload.UseProxy {
		h.Repository.UseProxy = true
	}

	h.Repository.Flush(payload.Symbol)
	h.SymbolsRepository.Slippage(payload.Symbol)

	return nil
}
