package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type OrdersTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.OrdersRepository
}

func (t *OrdersTask) Open() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Open(symbol)
	}
	return nil
}

func (t *OrdersTask) Sync() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Sync(symbol, 20)
	}
	return nil
}
