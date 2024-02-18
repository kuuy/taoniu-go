package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type AccountTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Account
  Repository  *repositories.AccountRepository
}

func NewAccountTask(ansqContext *common.AnsqClientContext) *AccountTask {
  return &AccountTask{
    AnsqContext: ansqContext,
    Repository: &repositories.AccountRepository{
      Db:  ansqContext.Db,
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
    },
  }
}

func (t *AccountTask) Flush() error {
  task, err := t.Job.Flush()
  if err != nil {
    return err
  }
  t.AnsqContext.Conn.Enqueue(
    task,
    asynq.Queue(config.ASYNQ_QUEUE_ACCOUNT),
    asynq.MaxRetry(0),
    asynq.Timeout(5*time.Minute),
  )
  return nil
}
