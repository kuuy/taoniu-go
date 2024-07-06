package margin

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/margin/cross"
)

type CrossTask struct {
  AnsqContext   *common.AnsqClientContext
  AccountTask   *tasks.AccountTask
  OrdersTask    *tasks.OrdersTask
  PositionsTask *tasks.PositionsTask
  TradingsTask  *tasks.TradingsTask
  AnalysisTask  *tasks.AnalysisTask
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

func (t *CrossTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = tasks.NewOrdersTask(t.AnsqContext)
  }
  return t.OrdersTask
}

func (t *CrossTask) Positions() *tasks.PositionsTask {
  if t.PositionsTask == nil {
    t.PositionsTask = tasks.NewPositionsTask(t.AnsqContext)
  }
  return t.PositionsTask
}

func (t *CrossTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = tasks.NewTradingsTask(t.AnsqContext)
  }
  return t.TradingsTask
}

func (t *CrossTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = tasks.NewAnalysisTask(t.AnsqContext)
  }
  return t.AnalysisTask
}
