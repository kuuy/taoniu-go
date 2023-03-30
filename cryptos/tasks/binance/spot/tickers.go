package spot

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"

	config "taoniu.local/cryptos/config/queue"
	jobs "taoniu.local/cryptos/queue/jobs/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersTask struct {
	Rdb               *redis.Client
	Ctx               context.Context
	Asynq             *asynq.Client
	Job               *jobs.Tickers
	Repository        *repositories.TickersRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func (t *TickersTask) Flush() error {
	symbols := t.SymbolsRepository.Scan()
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols) {
			j = len(symbols)
		}
		task, err := t.Job.Flush(symbols[i:j], false)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_TICKERS),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Minute),
		)
	}

	return nil
}

func (t *TickersTask) FlushDelay() error {
	symbols := t.SymbolsRepository.Symbols()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols) {
			j = len(symbols)
		}
		task, err := t.Job.Flush(symbols[i:j], true)
		if err != nil {
			return err
		}
		t.Asynq.Enqueue(
			task,
			asynq.Queue(config.BINANCE_SPOT_TICKERS_DELAY),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Minute),
		)
	}

	return nil
}
