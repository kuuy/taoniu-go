package spot

import (
  "taoniu.local/cryptos/common"
)

type CronTask struct {
  AnsqContext *common.AnsqClientContext
  SymbolsTask *SymbolsTask
  GridsTask   *GridsTask
  MarginTask  *MarginTask
}

func NewCronTask(ansqContext *common.AnsqClientContext) *CronTask {
  return &CronTask{
    AnsqContext: ansqContext,
  }
}

func (t *CronTask) Symbols() *SymbolsTask {
  if t.SymbolsTask == nil {
    t.SymbolsTask = NewSymbolsTask(t.AnsqContext)
  }
  return t.SymbolsTask
}

func (t *CronTask) Grids() *GridsTask {
  if t.GridsTask == nil {
    t.GridsTask = NewGridsTask(t.AnsqContext)
  }
  return t.GridsTask
}

func (t *CronTask) Margin() *MarginTask {
  if t.MarginTask == nil {
    t.MarginTask = NewMarginTask(t.AnsqContext)
  }
  return t.MarginTask
}

func (t *CronTask) Hourly() error {
  t.Symbols().Flush()
  t.Symbols().Count()
  //t.Grids().Flush()
  //t.Margin().Sync()

  return nil
}
