package tradings

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type GridsTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.GridsRepository
}

func (t *GridsTask) Flush() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Flush(symbol)
	}
	return nil
}

func (t *GridsTask) Update() error {
	return t.Repository.Update()
}
