package dydx

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx"
)

type PlansTask struct {
  Asynq *asynq.Client
  Job   *jobs.Plans
}

func (t *PlansTask) Flush(interval string) error {
  task, err := t.Job.Flush(interval)
  if err != nil {
    return err
  }
  t.Asynq.Enqueue(
    task,
    asynq.Queue(config.DYDX_PLANS),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}
