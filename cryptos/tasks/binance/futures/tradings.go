package futures

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/futures/tradings"
)

type TradingsTask struct {
  AnsqContext  *common.AnsqClientContext
  TriggersTask *tasks.TriggersTask
  ScalpingTask *tasks.ScalpingTask
  GamblingTask *tasks.GamblingTask
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

func (t *TradingsTask) Gambling() *tasks.GamblingTask {
  if t.GamblingTask == nil {
    t.GamblingTask = tasks.NewGamblingTask(t.AnsqContext)
  }
  return t.GamblingTask
}
