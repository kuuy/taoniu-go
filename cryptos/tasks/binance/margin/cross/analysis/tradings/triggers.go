package tradings

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross/analysis/tradings"
)

type TriggersTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.TriggersRepository
}

func NewTriggersTask(ansqContext *common.AnsqClientContext) *TriggersTask {
  return &TriggersTask{
    AnsqContext: ansqContext,
    Repository: &repositories.TriggersRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *TriggersTask) Flush() (err error) {
  t.Repository.Flush()
  return
}
