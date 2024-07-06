package binance

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot"
)

type SpotTask struct {
  AnsqContext    *common.AnsqClientContext
  SymbolsTask    *tasks.SymbolsTask
  TickersTask    *tasks.TickersTask
  DepthTask      *tasks.DepthTask
  KlinesTask     *tasks.KlinesTask
  IndicatorsTask *tasks.IndicatorsTask
  StrategiesTask *tasks.StrategiesTask
  PlansTask      *tasks.PlansTask
  TradingsTask   *tasks.TradingsTask
  AccountTask    *tasks.AccountTask
  OrdersTask     *tasks.OrdersTask
  PositionsTask  *tasks.PositionsTask
  AnalysisTask   *tasks.AnalysisTask
}

func NewSpotTask(ansqContext *common.AnsqClientContext) *SpotTask {
  return &SpotTask{
    AnsqContext: ansqContext,
  }
}

func (t *SpotTask) Account() *tasks.AccountTask {
  if t.AccountTask == nil {
    t.AccountTask = tasks.NewAccountTask(t.AnsqContext)
  }
  return t.AccountTask
}

func (t *SpotTask) Symbols() *tasks.SymbolsTask {
  if t.SymbolsTask == nil {
    t.SymbolsTask = tasks.NewSymbolsTask(t.AnsqContext)
  }
  return t.SymbolsTask
}

func (t *SpotTask) Tickers() *tasks.TickersTask {
  if t.TickersTask == nil {
    t.TickersTask = tasks.NewTickersTask(t.AnsqContext)
  }
  return t.TickersTask
}

func (t *SpotTask) Klines() *tasks.KlinesTask {
  if t.KlinesTask == nil {
    t.KlinesTask = tasks.NewKlinesTask(t.AnsqContext)
  }
  return t.KlinesTask
}

func (t *SpotTask) Depth() *tasks.DepthTask {
  if t.DepthTask == nil {
    t.DepthTask = tasks.NewDepthTask(t.AnsqContext)
  }
  return t.DepthTask
}

func (t *SpotTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = tasks.NewOrdersTask(t.AnsqContext)
  }
  return t.OrdersTask
}

func (t *SpotTask) Positions() *tasks.PositionsTask {
  if t.PositionsTask == nil {
    t.PositionsTask = tasks.NewPositionsTask(t.AnsqContext)
  }
  return t.PositionsTask
}

func (t *SpotTask) Indicators() *tasks.IndicatorsTask {
  if t.IndicatorsTask == nil {
    t.IndicatorsTask = tasks.NewIndicatorsTask(t.AnsqContext)
  }
  return t.IndicatorsTask
}

func (t *SpotTask) Strategies() *tasks.StrategiesTask {
  if t.StrategiesTask == nil {
    t.StrategiesTask = tasks.NewStrategiesTask(t.AnsqContext)
  }
  return t.StrategiesTask
}

func (t *SpotTask) Plans() *tasks.PlansTask {
  if t.PlansTask == nil {
    t.PlansTask = tasks.NewPlansTask(t.AnsqContext)
  }
  return t.PlansTask
}

func (t *SpotTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = tasks.NewTradingsTask(t.AnsqContext)
  }
  return t.TradingsTask
}

func (t *SpotTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = tasks.NewAnalysisTask(t.AnsqContext)
  }
  return t.AnalysisTask
}
