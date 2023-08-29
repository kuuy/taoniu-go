package tradings

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx/tradings"
  repositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type TriggersTask struct {
  Asynq      *asynq.Client
  Job        *jobs.Triggers
  Repository *repositories.TriggersRepository
}

func (t *TriggersTask) Place() error {
  ids := t.Repository.Ids()
  for _, id := range ids {
    task, err := t.Job.Place(id)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *TriggersTask) Flush() error {
  ids := t.Repository.TriggerIds()
  for _, id := range ids {
    task, err := t.Job.Flush(id)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
