package tradings

import (
  "taoniu.local/cryptos/common"
  "time"

  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/margin/cross/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
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

func (t *TriggersTask) Flush() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_MARGIN_CROSS_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *TriggersTask) Place() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Place(symbol)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_MARGIN_CROSS_TRADINGS_TRIGGERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
