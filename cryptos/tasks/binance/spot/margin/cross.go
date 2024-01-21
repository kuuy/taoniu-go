package margin

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/cross"
)

type CrossTask struct {
  AnsqContext  *common.AnsqClientContext
  AccountTask  *tasks.AccountTask
  TradingsTask *tasks.TradingsTask
}

func NewCrossTask(ansqContext *common.AnsqClientContext) *CrossTask {
  return &CrossTask{
    AnsqContext: ansqContext,
  }
}

func (t *CrossTask) Account() *tasks.AccountTask {
  if t.AccountTask == nil {
    t.AccountTask = tasks.NewAccountTask(t.AnsqContext)
  }
  return t.AccountTask
}

func (t *CrossTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = tasks.NewTradingsTask(t.AnsqContext)
  }
  return t.TradingsTask
}
