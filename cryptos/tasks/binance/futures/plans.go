package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type PlansTask struct {
  AnsqContext       *common.AnsqClientContext
  Job               *jobs.Plans
  Repository        *repositories.StrategiesRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewPlansTask(ansqContext *common.AnsqClientContext) *PlansTask {
  return &PlansTask{
    AnsqContext: ansqContext,
    Repository: &repositories.StrategiesRepository{
      Db: ansqContext.Db,
    },
    SymbolsRepository: &repositories.SymbolsRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *PlansTask) Flush(interval string) error {
  task, err := t.Job.Flush(interval)
  if err != nil {
    return err
  }
  t.AnsqContext.Conn.Enqueue(
    task,
    asynq.Queue(config.ASYNQ_QUEUE_PLANS),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}

func (t *PlansTask) Clean() error {
  symbols := t.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    t.Repository.Clean(symbol)
  }
  return nil
}
