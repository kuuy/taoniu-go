package tradingview

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/tradingview"
)

type AnalysisTask struct {
	Rdb              *redis.Client
	Ctx              context.Context
	Repository       *repositories.AnalysisRepository
	SymbolRepository *spotRepositories.SymbolsRepository
}

func (t *AnalysisTask) Flush() error {
	mutex := common.NewMutex(
		t.Rdb,
		t.Ctx,
		"locks:tradingview:analysis:flush",
	)
	if mutex.Lock(10 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	symbols := t.SymbolRepository.Scan()
	for _, symbol := range symbols {
		err := t.Repository.Flush("BINANCE", symbol, "1m")
		if err != nil {
			log.Println("analysis flush error", err)
		}
	}
	return nil
}
