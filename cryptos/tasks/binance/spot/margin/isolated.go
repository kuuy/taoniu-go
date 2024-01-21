package margin

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated"
)

type IsolatedTask struct {
  AnsqContext  *common.AnsqClientContext
  SymbolsTask  *tasks.SymbolsTask
  AccountTask  *tasks.AccountTask
  OrdersTask   *tasks.OrdersTask
  TradingsTask *tasks.TradingsTask
}

func NewIsolatedTask(ansqContext *common.AnsqClientContext) *IsolatedTask {
  return &IsolatedTask{
    AnsqContext: ansqContext,
  }
}

func (t *IsolatedTask) Symbols() *tasks.SymbolsTask {
  if t.SymbolsTask == nil {
    t.SymbolsTask = tasks.NewSymbolsTask(t.AnsqContext)
  }
  return t.SymbolsTask
}

func (t *IsolatedTask) Account() *tasks.AccountTask {
  if t.AccountTask == nil {
    t.AccountTask = tasks.NewAccountTask(t.AnsqContext)
  }
  return t.AccountTask
}

func (t *IsolatedTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = tasks.NewOrdersTask(t.AnsqContext)
  }
  return t.OrdersTask
}

func (t *IsolatedTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = tasks.NewTradingsTask(t.AnsqContext)
  }
  return t.TradingsTask
}
