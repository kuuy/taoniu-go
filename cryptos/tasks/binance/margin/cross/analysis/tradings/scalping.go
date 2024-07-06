package tradings

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross/analysis/tradings"
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

func (t *ScalpingTask) Flush() (err error) {
  t.Repository.Flush()
  return
}
