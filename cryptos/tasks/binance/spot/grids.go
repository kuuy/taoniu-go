package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type GridsTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.GridsRepository
}

func (t *GridsTask) Flush() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:grids:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Flush(symbol)
	}
	return nil
}
