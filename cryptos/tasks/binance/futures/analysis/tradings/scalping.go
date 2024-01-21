package tradings

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
)

type ScalpingTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.ScalpingRepository
}

func NewScalpingTask(ansqContext *common.AnsqClientContext) *ScalpingTask {
  return &ScalpingTask{
    AnsqContext: ansqContext,
    Repository: &repositories.ScalpingRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *ScalpingTask) Flush() error {
  t.Repository.Flush(1)
  t.Repository.Flush(2)
  return nil
}
