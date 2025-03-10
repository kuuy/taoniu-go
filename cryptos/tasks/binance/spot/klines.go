package spot

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Klines
  Repository         *repositories.KlinesRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewKlinesTask(ansqContext *common.AnsqClientContext) *KlinesTask {
  return &KlinesTask{
    AnsqContext: ansqContext,
    Repository: &repositories.KlinesRepository{
      Db: ansqContext.Db,
    },
    ScalpingRepository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *KlinesTask) Clean() error {
  for _, symbol := range t.ScalpingRepository.Scan() {
    task, err := t.Job.Clean(symbol)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_KLINES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}
