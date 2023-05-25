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

type Scalping struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	Repository        *repositories.ScalpingRepository
	SymbolsRepository *spotRepositories.SymbolsRepository
}

func NewScalping() *Scalping {
	h := &Scalping{
		Db:  common.NewDB(),
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.ScalpingRepository{
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

type ScalpingFlushPayload struct {
	Symbol string
}

func (h *Scalping) Flush(ctx context.Context, t *asynq.Task) error {
	var payload ScalpingFlushPayload
	json.Unmarshal(t.Payload(), &payload)

	h.Repository.Flush(payload.Symbol)

	return nil
}
