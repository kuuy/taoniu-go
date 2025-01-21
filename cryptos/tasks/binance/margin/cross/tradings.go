package cross

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/margin/cross/tradings"
)

type TradingsTask struct {
  AnsqContext  *common.AnsqClientContext
  ScalpingTask *tasks.ScalpingTask
}

func NewTradingsTask(ansqContext *common.AnsqClientContext) *TradingsTask {
  return &TradingsTask{
    AnsqContext: ansqContext,
  }
}

func (t *TradingsTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = tasks.NewScalpingTask(t.AnsqContext)
  }
  return t.ScalpingTask
}
