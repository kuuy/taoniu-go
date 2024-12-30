package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type DepthTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Depth
  ScalpingRepository *repositories.ScalpingRepository
}

func NewDepthTask(ansqContext *common.AnsqClientContext) *DepthTask {
  return &DepthTask{
    AnsqContext: ansqContext,
    ScalpingRepository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *DepthTask) Flush(limit int) error {
  for _, symbol := range t.ScalpingRepository.Scan(2) {
    task, err := t.Job.Flush(symbol, limit, false)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_DEPTH),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
