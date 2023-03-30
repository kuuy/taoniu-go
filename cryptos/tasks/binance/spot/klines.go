package spot

import (
	"math/rand"
	"time"

	"github.com/hibiken/asynq"

	config "taoniu.local/cryptos/config/queue"
	jobs "taoniu.local/cryptos/queue/jobs/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesTask struct {
	Asynq             *asynq.Client
	Job               *jobs.Klines
	Repository        *repositories.KlinesRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func (t *KlinesTask) Flush(interval string, limit int) error {
	symbols := t.SymbolsRepository.Scan()
	for _, symbol := range symbols {
		task, err := t.Job.Flush(symbol, interval, limit, false)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_KLINES),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Second),
		)
		t.Repository.Flush(symbol, interval, limit)
	}
	return nil
}

func (t *KlinesTask) FlushDelay(interval string, limit int) error {
	symbols := t.SymbolsRepository.Symbols()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
	for _, symbol := range symbols {
		task, err := t.Job.Flush(symbol, interval, limit, true)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_KLINES_DELAY),
			asynq.MaxRetry(0),
			asynq.Timeout(8*time.Second),
		)
	}
	return nil
}

func (t *KlinesTask) Clean() error {
	t.Repository.Clean()
	return nil
}
