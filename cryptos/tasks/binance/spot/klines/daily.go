package klines

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/klines"
)

type DailyTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Flush(limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Flush(symbol, limit)
	}
	return nil
}
