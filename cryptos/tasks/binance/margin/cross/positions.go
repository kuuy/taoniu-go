package cross

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/margin/cross"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/margin/cross"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross"
)

type PositionsTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Positions
  Repository         *repositories.PositionsRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewPositionsTask(ansqContext *common.AnsqClientContext) *PositionsTask {
  return &PositionsTask{
    AnsqContext: ansqContext,
    Repository: &repositories.PositionsRepository{
      Db: ansqContext.Db,
    },
    ScalpingRepository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *PositionsTask) Flush() error {
  for symbol, side := range t.ScalpingRepository.Scan() {
    task, err := t.Job.Flush(symbol, side)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_POSITIONS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
