package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type DepthTask struct {
	Rdb               *redis.Client
	Ctx               context.Context
	Repository        *repositories.DepthRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func (t *DepthTask) Flush() error {
	mutex := common.NewMutex(
		t.Rdb,
		t.Ctx,
		"locks:binance:spot:depth:flush",
	)
	if mutex.Lock(10 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	symbols := t.SymbolsRepository.Scan()
	for _, symbol := range symbols {
		t.Repository.Flush(symbol)
		t.SymbolsRepository.Slippage(symbol)
	}

	return nil
}
