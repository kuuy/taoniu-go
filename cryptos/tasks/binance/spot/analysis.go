package spot

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/analysis"
)

type AnalysisTask struct {
  AnsqContext  *common.AnsqClientContext
  TradingsTask *tasks.TradingsTask
  MarginTask   *tasks.MarginTask
}

func NewAnalysisTask(ansqContext *common.AnsqClientContext) *AnalysisTask {
  return &AnalysisTask{
    AnsqContext: ansqContext,
  }
}

func (t *AnalysisTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = &tasks.TradingsTask{
      Db: t.AnsqContext.Db,
    }
  }
  return t.TradingsTask
}

func (t *AnalysisTask) Margin() *tasks.MarginTask {
  if t.MarginTask == nil {
    t.MarginTask = &tasks.MarginTask{
      Db: t.AnsqContext.Db,
    }
  }
  return t.MarginTask
}

func (t *AnalysisTask) Flush() {
  t.Tradings().Flush()
  t.Margin().Flush()
}
