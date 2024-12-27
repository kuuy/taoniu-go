package gambling

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/tradings/gambling"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings/gambling"
)

type AntTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Ant
  Repository  *repositories.AntRepository
}

func NewAntTask(ansqContext *common.AnsqClientContext) *AntTask {
  return &AntTask{
    AnsqContext: ansqContext,
    Repository: &repositories.AntRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *AntTask) Place() error {
  ids := t.Repository.Ids()
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

func (t *AntTask) Flush() error {
  ids := t.Repository.Ids()
  for _, id := range ids {
    task, err := t.Job.Flush(id)
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
