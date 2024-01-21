package tradings

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/margin/isolated/tradings"
  "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated/tradings/fishers"
)

type FishersTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Fishers
  Repository  *tradings.FishersRepository
  GridsTask   *tasks.GridsTask
}

func NewFishersTask(ansqContext *common.AnsqClientContext) *FishersTask {
  return &FishersTask{
    AnsqContext: ansqContext,
    Repository: &tradings.FishersRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *FishersTask) Grids() *tasks.GridsTask {
  if t.GridsTask == nil {
    t.GridsTask = tasks.NewGridsTask(t.AnsqContext)
  }
  return t.GridsTask
}

func (t *FishersTask) Flush() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_MARGIN_ISOLATED_TRADINGS_FISHERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *FishersTask) Place() error {
  symbols := t.Repository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Place(symbol)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_MARGIN_ISOLATED_TRADINGS_FISHERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
