package cross

import (
  "taoniu.local/cryptos/common"

  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/cross/tradings"
)

type TradingsTask struct {
  AnsqContext  *common.AnsqClientContext
  TriggersTask *tasks.TriggersTask
}

func NewTradingsTask(ansqContext *common.AnsqClientContext) *TradingsTask {
  return &TradingsTask{
    AnsqContext: ansqContext,
  }
}

func (t *TradingsTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = tasks.NewTriggersTask(t.AnsqContext)
  }
  return t.TriggersTask
}
