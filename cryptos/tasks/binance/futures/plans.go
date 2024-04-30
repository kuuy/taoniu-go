package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type PlansTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Plans
  Repository         *repositories.PlansRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewPlansTask(ansqContext *common.AnsqClientContext) *PlansTask {
  return &PlansTask{
    AnsqContext: ansqContext,
    Repository: &repositories.PlansRepository{
      Db: ansqContext.Db,
    },
    TradingsRepository: &repositories.TradingsRepository{
      Db: ansqContext.Db,
      ScalpingRepository: &tradingsRepositories.ScalpingRepository{
        Db: ansqContext.Db,
      },
      TriggersRepository: &tradingsRepositories.TriggersRepository{
        Db: ansqContext.Db,
      },
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
  for _, symbol := range t.TradingsRepository.Scan() {
    t.Repository.Clean(symbol)
  }
  return nil
}
