package tasks

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance"
)

type BinanceTask struct {
  AnsqContext *common.AnsqClientContext
  SpotTask    *tasks.SpotTask
  MarginTask  *tasks.MarginTask
  FuturesTask *tasks.FuturesTask
  SavingsTask *tasks.SavingsTask
  ServerTask  *tasks.ServerTask
}

func NewBinanceTask(ansqContext *common.AnsqClientContext) *BinanceTask {
  return &BinanceTask{
    AnsqContext: ansqContext,
  }
}

func (t *BinanceTask) Spot() *tasks.SpotTask {
  if t.SpotTask == nil {
    t.SpotTask = tasks.NewSpotTask(t.AnsqContext)
  }
  return t.SpotTask
}

func (t *BinanceTask) Margin() *tasks.MarginTask {
  if t.MarginTask == nil {
    t.MarginTask = tasks.NewMarginTask(t.AnsqContext)
  }
  return t.MarginTask
}

func (t *BinanceTask) Futures() *tasks.FuturesTask {
  if t.FuturesTask == nil {
    t.FuturesTask = tasks.NewFuturesTask(t.AnsqContext)
  }
  return t.FuturesTask
}

func (t *BinanceTask) Savings() *tasks.SavingsTask {
  if t.SavingsTask == nil {
    t.SavingsTask = tasks.NewSavingsTask(t.AnsqContext)
  }
  return t.SavingsTask
}

func (t *BinanceTask) Server() *tasks.ServerTask {
  if t.ServerTask == nil {
    t.ServerTask = tasks.NewServerTask(t.AnsqContext)
  }
  return t.ServerTask
}
