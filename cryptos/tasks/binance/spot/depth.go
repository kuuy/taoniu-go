package spot

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"

	"taoniu.local/cryptos/common"
	config "taoniu.local/cryptos/config/queue"
	jobs "taoniu.local/cryptos/queue/jobs/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type DepthTask struct {
	Rdb               *redis.Client
	Ctx               context.Context
	Asynq             *asynq.Client
	Job               *jobs.Depth
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

func (t *DepthTask) FlushDelay() error {
	symbols := t.SymbolsRepository.Symbols()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
	for _, symbol := range symbols {
		task, err := t.Job.Flush(symbol)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_DEPTH),
			asynq.MaxRetry(0),
			asynq.Timeout(10*time.Second),
		)
	}

	return nil
}
