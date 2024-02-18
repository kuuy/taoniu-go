package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
)

type PlansTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Plans
}

func NewPlansTask(ansqContext *common.AnsqClientContext) *PlansTask {
  return &PlansTask{
    AnsqContext: ansqContext,
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
