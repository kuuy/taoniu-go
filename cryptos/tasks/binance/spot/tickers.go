package spot

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.TickersRepository
}

func (t *TickersTask) Flush() error {
	mutex := pool.NewMutex(
		t.Rdb,
		t.Ctx,
		"locks:binance:spot:tickers:flush",
	)
	if mutex.Lock(10 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	symbols, _ := t.Rdb.ZRevRange(
		t.Ctx,
		"binance:spot:tickers:flush",
		0,
		-1,
	).Result()
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols) {
			j = len(symbols)
		}
		t.Repository.Flush(symbols[i:j])
	}

	return nil
}
