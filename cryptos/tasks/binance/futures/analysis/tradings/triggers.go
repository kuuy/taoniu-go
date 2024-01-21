package tradings

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
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

func (t *TriggersTask) Flush() error {
  t.Repository.Flush(1)
  t.Repository.Flush(2)
  return nil
}
