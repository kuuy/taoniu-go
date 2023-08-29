package dydx

import (
  "context"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type OrderbookTask struct {
  Rdb                *redis.Client
  Ctx                context.Context
  Asynq              *asynq.Client
  Job                *jobs.Orderbook
  MarketsRepository  *repositories.MarketsRepository
  TradingsRepository *repositories.TradingsRepository
}

func (t *OrderbookTask) Flush() error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_ORDERBOOK),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *OrderbookTask) FlushDelay() error {
  symbols := t.MarketsRepository.Symbols()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_ORDERBOOK_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *OrderbookTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (t *OrderbookTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
