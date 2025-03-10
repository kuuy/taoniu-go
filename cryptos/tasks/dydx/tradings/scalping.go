package tradings

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx/tradings"
  dydxRepositories "taoniu.local/cryptos/repositories/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type ScalpingTask struct {
  Asynq            *asynq.Client
  Job              *jobs.Scalping
  Repository       *repositories.ScalpingRepository
  ParentRepository *dydxRepositories.ScalpingRepository
}

func (t *ScalpingTask) Place() error {
  ids := t.ParentRepository.PlanIds(0)
  for _, id := range ids {
    task, err := t.Job.Place(id)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_TRADINGS_SCALPING),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *ScalpingTask) Flush() error {
  ids := t.Repository.ScalpingIds()
  for _, id := range ids {
    task, err := t.Job.Flush(id)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_TRADINGS_SCALPING),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
