package tradings

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/margin/cross"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/margin/cross/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross/tradings"
)

type TriggersTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Triggers
  Repository  *repositories.TriggersRepository
}

func NewTriggersTask(ansqContext *common.AnsqClientContext) *TriggersTask {
  return &TriggersTask{
    AnsqContext: ansqContext,
    Repository: &repositories.TriggersRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *TriggersTask) Place() error {
  ids := t.Repository.Ids()
  for _, id := range ids {
    task, err := t.Job.Place(id)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *TriggersTask) Flush() error {
  ids := t.Repository.TriggerIds()
  for _, id := range ids {
    task, err := t.Job.Flush(id)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
