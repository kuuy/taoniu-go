package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type TradingsTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.TradingsRepository
}

func (t *TradingsTask) Grids() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Grids(symbol)
	}

	return nil
}
