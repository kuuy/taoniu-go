package analysis

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/tradings"
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

func (t *TradingsTask) Flush() error {
  t.Scalping().Flush()
  return nil
}
