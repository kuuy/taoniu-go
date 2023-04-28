package tradings

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type Triggers struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	Repository        *repositories.TriggersRepository
	SymbolsRepository *spotRepositories.SymbolsRepository
}

func NewTriggers() *Triggers {
	h := &Triggers{
		Db:  common.NewDB(),
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.TriggersRepository{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}
	h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}
	h.Repository.AccountRepository = &spotRepositories.AccountRepository{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}
	h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
		Db:  h.Db,
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}

	return h
}

type TriggersPlacePayload struct {
	Symbol string
}

type TriggersFlushPayload struct {
	Symbol string
}

func (h *Triggers) Place(ctx context.Context, t *asynq.Task) error {
	var payload TriggersPlacePayload
	json.Unmarshal(t.Payload(), &payload)

	h.Repository.Place(payload.Symbol)

	return nil
}

func (h *Triggers) Flush(ctx context.Context, t *asynq.Task) error {
	var payload TriggersFlushPayload
	json.Unmarshal(t.Payload(), &payload)

	h.Repository.Flush(payload.Symbol)

	return nil
}
