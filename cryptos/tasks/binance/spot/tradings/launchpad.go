package tradings

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type LaunchpadTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Launchpad
  Repository  *repositories.LaunchpadRepository
}

func NewLaunchpadTask(ansqContext *common.AnsqClientContext) *LaunchpadTask {
  return &LaunchpadTask{
    AnsqContext: ansqContext,
    Repository: &repositories.LaunchpadRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *LaunchpadTask) Place() error {
  ids := t.Repository.Ids()
  for _, id := range ids {
    task, err := t.Job.Place(id)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_LAUNCHPAD),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *LaunchpadTask) Flush() error {
  ids := t.Repository.LaunchpadIds()
  for _, id := range ids {
    task, err := t.Job.Flush(id)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_LAUNCHPAD),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
