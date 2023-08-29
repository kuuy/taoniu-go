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

type KlinesTask struct {
  Rdb                *redis.Client
  Ctx                context.Context
  Asynq              *asynq.Client
  Job                *jobs.Klines
  Repository         *repositories.KlinesRepository
  MarketsRepository  *repositories.MarketsRepository
  TradingsRepository *repositories.TradingsRepository
}

func (t *KlinesTask) Flush(interval string, limit int) error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, interval, 0, limit, false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_KLINES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *KlinesTask) FlushDelay(interval string, limit int) error {
  symbols := t.MarketsRepository.Symbols()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, interval, 0, limit, true)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_KLINES_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *KlinesTask) Clean() error {
  t.Repository.Clean()
  return nil
}

func (t *KlinesTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (t *KlinesTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
