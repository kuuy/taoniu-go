package tradings

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type TriggersTask struct {
  Asynq      *asynq.Client
  Job        *jobs.Triggers
  Repository *repositories.TriggersRepository
}

func (t *TriggersTask) Place() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Place(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *TriggersTask) Flush() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
