package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.KlinesRepository
}

func (t *KlinesTask) Flush(interval string, limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Flush(symbol, interval, limit)
	}
	return nil
}

func (t *KlinesTask) Clean() error {
	t.Repository.Clean()
	return nil
}
