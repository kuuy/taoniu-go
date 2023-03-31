package tradingview

import (
	"context"
	"math/rand"
	config "taoniu.local/cryptos/config/queue"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"

	jobs "taoniu.local/cryptos/queue/jobs/tradingview"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/tradingview"
)

type AnalysisTask struct {
	Rdb               *redis.Client
	Ctx               context.Context
	Asynq             *asynq.Client
	Job               *jobs.Analysis
	Repository        *repositories.AnalysisRepository
	SymbolsRepository *spotRepositories.SymbolsRepository
}

func (t *AnalysisTask) Flush() error {
	symbols := t.SymbolsRepository.Scan()
	for _, symbol := range symbols {
		task, err := t.Job.Flush("BINANCE", symbol, "1m", false)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.TRADINGVIEW_ANALYSIS),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Minute),
		)
	}
	return nil
}

func (t *AnalysisTask) FlushDelay() error {
	symbols := t.SymbolsRepository.Symbols()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
	for _, symbol := range symbols {
		task, err := t.Job.Flush("BINANCE", symbol, "1m", true)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.TRADINGVIEW_ANALYSIS_DELAY),
			asynq.MaxRetry(0),
			asynq.Timeout(10*time.Minute),
		)
	}
	return nil
}
