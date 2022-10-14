package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/daily/margin/isolated"
)

type ProfitsTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.ProfitsRepository
}

func (t *ProfitsTask) Flush() error {
	symbols, _ := t.Rdb.SMembers(t.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		t.Repository.Flush(symbol)
	}
	return nil
}
