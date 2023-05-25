package tradings

import (
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/tradings"
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type ScalpingTask struct {
  Asynq           *asynq.Client
  Job             *jobs.Scalping
  Repository      *repositories.ScalpingRepository
  PlansRepository *plansRepositories.DailyRepository
}

func (t *ScalpingTask) Place() error {
  plan, err := t.PlansRepository.Filter()
  if err != nil {
    return err
  }
  return t.Repository.Place(plan)
}

func (t *ScalpingTask) Flush() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_TRADINGS_SCALPING),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
