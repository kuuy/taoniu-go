package spot

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type PlansTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Plans
  Repository         *repositories.PlansRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewPlansTask(ansqContext *common.AnsqClientContext) *PlansTask {
  return &PlansTask{
    AnsqContext: ansqContext,
    Repository: &repositories.PlansRepository{
      Db: ansqContext.Db,
    },
    ScalpingRepository: &repositories.ScalpingRepository{
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
  for _, symbol := range t.ScalpingRepository.Scan() {
    t.Repository.Clean(symbol)
  }
  return nil
}
