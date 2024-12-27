package gambling

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/tradings/gambling"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type ScalpingTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Scalping
  Repository  *repositories.ScalpingRepository
}

func NewScalpingTask(ansqContext *common.AnsqClientContext) *ScalpingTask {
  return &ScalpingTask{
    AnsqContext: ansqContext,
    Repository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *ScalpingTask) Place() error {
  ids := t.Repository.ScalpingIds()
  for _, id := range ids {
    task, err := t.Job.Place(id)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_GAMBLING),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
