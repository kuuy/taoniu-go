package spot

import (
	"math/rand"
	"time"

	"github.com/hibiken/asynq"

	config "taoniu.local/cryptos/config/queue"
	jobs "taoniu.local/cryptos/queue/jobs/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type DepthTask struct {
	Asynq             *asynq.Client
	Job               *jobs.Depth
	Repository        *repositories.DepthRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func (t *DepthTask) Flush() error {
	symbols := t.SymbolsRepository.Scan()
	for _, symbol := range symbols {
		task, err := t.Job.Flush(symbol, false)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_DEPTH),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Second),
		)
	}

	return nil
}

func (t *DepthTask) FlushDelay() error {
	symbols := t.SymbolsRepository.Symbols()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
	for _, symbol := range symbols {
		task, err := t.Job.Flush(symbol, true)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_DEPTH_DELAY),
			asynq.MaxRetry(0),
			asynq.Timeout(8*time.Second),
		)
	}

	return nil
}
