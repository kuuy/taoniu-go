package spot

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin"
)

type MarginTask struct {
  AnsqContext  *common.AnsqClientContext
  CrossTask    *tasks.CrossTask
  IsolatedTask *tasks.IsolatedTask
  OrdersTask   *tasks.OrdersTask
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

func (t *MarginTask) Isolated() *tasks.IsolatedTask {
  if t.IsolatedTask == nil {
    t.IsolatedTask = tasks.NewIsolatedTask(t.AnsqContext)
  }
  return t.IsolatedTask
}

func (t *MarginTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = tasks.NewOrdersTask(t.AnsqContext)
  }
  return t.OrdersTask
}

func (t *MarginTask) Flush() {
  t.Cross().Account().Flush()
  t.Isolated().Account().Flush()
  t.Isolated().Account().Liquidate()
  t.Isolated().Orders().Open()
  t.Isolated().Symbols().Flush()
  t.Orders().Flush()
}

func (t *MarginTask) Sync() {
  t.Isolated().Orders().Sync()
}
