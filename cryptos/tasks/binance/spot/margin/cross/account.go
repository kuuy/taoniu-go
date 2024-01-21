package cross

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
)

type AccountTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.AccountRepository
}

func NewAccountTask(ansqContext *common.AnsqClientContext) *AccountTask {
  return &AccountTask{
    AnsqContext: ansqContext,
    Repository: &repositories.AccountRepository{
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
    },
  }
}

func (t *AccountTask) Flush() error {
  return t.Repository.Flush()
}
