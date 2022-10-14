package strategies

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/strategies"
)

type DailyTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Atr() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Atr(symbol)
	}
	return nil
}

func (t *DailyTask) Zlema() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Zlema(symbol)
	}
	return nil
}

func (t *DailyTask) HaZlema() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.HaZlema(symbol)
	}
	return nil
}

func (t *DailyTask) Kdj() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Kdj(symbol)
	}
	return nil
}

func (t *DailyTask) BBands() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.BBands(symbol)
	}
	return nil
}

func (t *DailyTask) Flush() error {
	t.Atr()
	t.Zlema()
	t.HaZlema()
	t.Kdj()
	t.BBands()
	return nil
}
