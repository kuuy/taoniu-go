package analysis

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/margin/cross/analysis/tradings"
)

type TradingsTask struct {
  AnsqContext  *common.AnsqClientContext
  ScalpingTask *tasks.ScalpingTask
  TriggersTask *tasks.TriggersTask
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

func (t *TradingsTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = tasks.NewTriggersTask(t.AnsqContext)
  }
  return t.TriggersTask
}

func (t *TradingsTask) Flush() error {
  t.Scalping().Flush()
  t.Triggers().Flush()
  return nil
}
