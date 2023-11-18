package dydx

import (
  "context"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx"
)

type TickersTask struct {
  Rdb   *redis.Client
  Ctx   context.Context
  Asynq *asynq.Client
  Job   *jobs.Tickers
}

func (t *TickersTask) Flush() error {
  task, err := t.Job.Flush()
  if err != nil {
    return err
  }
  t.Asynq.Enqueue(
    task,
    asynq.Queue(config.DYDX_TICKERS),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}
