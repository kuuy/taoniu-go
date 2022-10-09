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

func (t *KlinesTask) FlushDaily(limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.FlushDaily(symbol, limit)
	}
	return nil
}
