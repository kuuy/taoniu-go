package tradings

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/tradings/gambling"
)

type GamblingTask struct {
  AnsqContext  *common.AnsqClientContext
  AntTask      *tasks.AntTask
  ScalpingTask *tasks.ScalpingTask
}

func NewGamblingTask(ansqContext *common.AnsqClientContext) *GamblingTask {
  return &GamblingTask{
    AnsqContext: ansqContext,
  }
}

func (t *GamblingTask) Ant() *tasks.AntTask {
  if t.AntTask == nil {
    t.AntTask = tasks.NewAntTask(t.AnsqContext)
  }
  return t.AntTask
}

func (t *GamblingTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = tasks.NewScalpingTask(t.AnsqContext)
  }
  return t.ScalpingTask
}
