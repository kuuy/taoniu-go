package spot

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type AccountTask struct {
  Asynq      *asynq.Client
  Job        *jobs.Account
  Repository *repositories.AccountRepository
}

func (t *AccountTask) Flush() error {
  task, err := t.Job.Flush()
  if err != nil {
    return err
  }
  t.Asynq.Enqueue(
    task,
    asynq.Queue(config.BINANCE_SPOT_ACCOUNT),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}
