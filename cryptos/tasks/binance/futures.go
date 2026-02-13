package binance

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/futures"
)

type FuturesTask struct {
  AnsqContext     *common.AnsqClientContext
  CronTask        *tasks.CronTask
  AccountTask     *tasks.AccountTask
  SymbolsTask     *tasks.SymbolsTask
  TickersTask     *tasks.TickersTask
  FundingRateTask *tasks.FundingRateTask
  KlinesTask      *tasks.KlinesTask
  DepthTask       *tasks.DepthTask
  PatternsTask    *tasks.PatternsTask
  OrdersTask      *tasks.OrdersTask
  IndicatorsTask  *tasks.IndicatorsTask
  StrategiesTask  *tasks.StrategiesTask
  PlansTask       *tasks.PlansTask
  ScalpingTask    *tasks.ScalpingTask
  TradingsTask    *tasks.TradingsTask
  AnalysisTask    *tasks.AnalysisTask
}

func NewFuturesTask(ansqContext *common.AnsqClientContext) *FuturesTask {
  return &FuturesTask{
    AnsqContext: ansqContext,
  }
}

func (t *FuturesTask) Cron() *tasks.CronTask {
  if t.CronTask == nil {
    t.CronTask = tasks.NewCronTask(t.AnsqContext)
  }
  return t.CronTask
}

func (t *FuturesTask) Account() *tasks.AccountTask {
  if t.AccountTask == nil {
    t.AccountTask = tasks.NewAccountTask(t.AnsqContext)
  }
  return t.AccountTask
}

func (t *FuturesTask) Symbols() *tasks.SymbolsTask {
  if t.SymbolsTask == nil {
    t.SymbolsTask = tasks.NewSymbolsTask(t.AnsqContext)
  }
  return t.SymbolsTask
}

func (t *FuturesTask) Tickers() *tasks.TickersTask {
  if t.TickersTask == nil {
    t.TickersTask = tasks.NewTickersTask(t.AnsqContext)
  }
  return t.TickersTask
}

func (t *FuturesTask) FundingRate() *tasks.FundingRateTask {
  if t.FundingRateTask == nil {
    t.FundingRateTask = tasks.NewFundingRateTask(t.AnsqContext)
  }
  return t.FundingRateTask
}

func (t *FuturesTask) Klines() *tasks.KlinesTask {
  if t.KlinesTask == nil {
    t.KlinesTask = tasks.NewKlinesTask(t.AnsqContext)
  }
  return t.KlinesTask
}

func (t *FuturesTask) Depth() *tasks.DepthTask {
  if t.DepthTask == nil {
    t.DepthTask = tasks.NewDepthTask(t.AnsqContext)
  }
  return t.DepthTask
}

func (t *FuturesTask) Patterns() *tasks.PatternsTask {
  if t.PatternsTask == nil {
    t.PatternsTask = tasks.NewPatternsTask(t.AnsqContext)
  }
  return t.PatternsTask
}

func (t *FuturesTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = tasks.NewOrdersTask(t.AnsqContext)
  }
  return t.OrdersTask
}

func (t *FuturesTask) Indicators() *tasks.IndicatorsTask {
  if t.IndicatorsTask == nil {
    t.IndicatorsTask = tasks.NewIndicatorsTask(t.AnsqContext)
  }
  return t.IndicatorsTask
}

func (t *FuturesTask) Strategies() *tasks.StrategiesTask {
  if t.StrategiesTask == nil {
    t.StrategiesTask = tasks.NewStrategiesTask(t.AnsqContext)
  }
  return t.StrategiesTask
}

func (t *FuturesTask) Plans() *tasks.PlansTask {
  if t.PlansTask == nil {
    t.PlansTask = tasks.NewPlansTask(t.AnsqContext)
  }
  return t.PlansTask
}

func (t *FuturesTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = tasks.NewScalpingTask(t.AnsqContext)
  }
  return t.ScalpingTask
}

func (t *FuturesTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = tasks.NewTradingsTask(t.AnsqContext)
  }
  return t.TradingsTask
}

func (t *FuturesTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = tasks.NewAnalysisTask(t.AnsqContext)
  }
  return t.AnalysisTask
}
