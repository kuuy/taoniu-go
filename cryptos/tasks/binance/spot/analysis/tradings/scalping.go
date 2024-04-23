package tradings

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings"
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
  t.Repository.Flush()
  return nil
}
