package spot

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type DepthTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Depth
  SymbolsRepository  *repositories.SymbolsRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewDepthTask(ansqContext *common.AnsqClientContext) *DepthTask {
  return &DepthTask{
    AnsqContext: ansqContext,
    SymbolsRepository: &repositories.SymbolsRepository{
      Db: ansqContext.Db,
    },
    ScalpingRepository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *DepthTask) Flush(limit int) error {
  symbols := t.ScalpingRepository.Scan()
  for _, symbol := range symbols {
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
