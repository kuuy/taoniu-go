package spot

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type PositionsTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Positions
  Repository         *repositories.PositionsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewPositionsTask(ansqContext *common.AnsqClientContext) *PositionsTask {
  return &PositionsTask{
    AnsqContext: ansqContext,
    Repository: &repositories.PositionsRepository{
      Db: ansqContext.Db,
    },
    TradingsRepository: &repositories.TradingsRepository{
      Db: ansqContext.Db,
      ScalpingRepository: &tradingsRepositories.ScalpingRepository{
        Db: ansqContext.Db,
      },
    },
  }
}

func (t *PositionsTask) Flush() error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Flush(symbol)
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
