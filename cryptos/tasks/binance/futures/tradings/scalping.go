package tradings

import (
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/tradings"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type ScalpingTask struct {
  AnsqContext      *common.AnsqClientContext
  Job              *jobs.Scalping
  Repository       *repositories.ScalpingRepository
  ParentRepository *futuresRepositories.ScalpingRepository
}

func NewScalpingTask(ansqContext *common.AnsqClientContext) *ScalpingTask {
  return &ScalpingTask{
    AnsqContext: ansqContext,
    Repository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
    ParentRepository: &futuresRepositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *ScalpingTask) Place() error {
  planIds := t.ParentRepository.PlanIds(0)
  for _, planId := range planIds {
    mutex := common.NewMutex(
      t.AnsqContext.Rdb,
      t.AnsqContext.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_PLACE, planId),
    )
    if !mutex.Lock(30 * time.Second) {
      continue
    }
    task, err := t.Job.Place(planId)
    if err != nil {
      continue
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_SCALPING),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *ScalpingTask) Flush() error {
  ids := t.Repository.ScalpingIds()
  for _, id := range ids {
    task, err := t.Job.Flush(id)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TRADINGS_SCALPING),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
