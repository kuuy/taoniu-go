package plans

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/plans"
)

type MinutelyTask struct {
  Asynq *asynq.Client
  Job   *jobs.Minutely
}

func (t *MinutelyTask) Flush() error {
  task, err := t.Job.Flush()
  if err != nil {
    return err
  }
  t.Asynq.Enqueue(
    task,
    asynq.Queue(config.BINANCE_SPOT_PLANS),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}
