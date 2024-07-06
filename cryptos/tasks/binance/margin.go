package binance

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/margin"
)

type MarginTask struct {
  AnsqContext *common.AnsqClientContext
  CrossTask   *tasks.CrossTask
}

func NewMarginTask(ansqContext *common.AnsqClientContext) *MarginTask {
  return &MarginTask{
    AnsqContext: ansqContext,
  }
}

func (t *MarginTask) Cross() *tasks.CrossTask {
  if t.CrossTask == nil {
    t.CrossTask = tasks.NewCrossTask(t.AnsqContext)
  }
  return t.CrossTask
}
