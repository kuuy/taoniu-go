package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type FundingRateTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.FundingRate
  ScalpingRepository *repositories.ScalpingRepository
}

func NewFundingRateTask(ansqContext *common.AnsqClientContext) *FundingRateTask {
  return &FundingRateTask{
    AnsqContext: ansqContext,
    ScalpingRepository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *FundingRateTask) Flush() error {
  task, err := t.Job.Flush()
  if err != nil {
    return err
  }
  t.AnsqContext.Conn.Enqueue(
    task,
    asynq.Queue(config.ASYNQ_QUEUE_FUNDING),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}
