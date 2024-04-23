package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type DepthTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Depth
  TradingsRepository *repositories.TradingsRepository
}

func NewDepthTask(ansqContext *common.AnsqClientContext) *DepthTask {
  return &DepthTask{
    AnsqContext: ansqContext,
    TradingsRepository: &repositories.TradingsRepository{
      Db: ansqContext.Db,
      ScalpingRepository: &tradingsRepositories.ScalpingRepository{
        Db: ansqContext.Db,
      },
      TriggersRepository: &tradingsRepositories.TriggersRepository{
        Db: ansqContext.Db,
      },
    },
  }
}

func (t *DepthTask) Flush(limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
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

func (t *DepthTask) FlushDelay(limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Flush(symbol, limit, true)
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
