package spot

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/scalping"
)

type ScalpingTask struct {
  AnsqContext *common.AnsqClientContext
  PlansTask   *tasks.PlansTask
}

func NewScalpingTask(ansqContext *common.AnsqClientContext) *ScalpingTask {
  return &ScalpingTask{
    AnsqContext: ansqContext,
  }
}

func (t *ScalpingTask) Plans() *tasks.PlansTask {
  if t.PlansTask == nil {
    t.PlansTask = tasks.NewPlansTask(t.AnsqContext)
  }
  return t.PlansTask
}
