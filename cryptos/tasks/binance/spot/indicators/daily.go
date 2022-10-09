package indicators

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type DailyTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Pivot() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Pivot(symbol)
	}
	return nil
}

func (t *DailyTask) Atr(period int, limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Atr(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) Zlema(period int, limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Zlema(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) HaZlema(period int, limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.HaZlema(symbol, period, limit)
	}
	return nil
}

func (t *DailyTask) Kdj(longPeriod int, shortPeriod int, limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Kdj(symbol, longPeriod, shortPeriod, limit)
	}
	return nil
}

func (t *DailyTask) BBands(period int, limit int) error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.BBands(symbol, period, limit)
	}
	return nil
}
