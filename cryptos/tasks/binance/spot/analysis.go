package spot

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/analysis"
)

type AnalysisTask struct {
  AnsqContext  *common.AnsqClientContext
  TradingsTask *tasks.TradingsTask
}

func NewAnalysisTask(ansqContext *common.AnsqClientContext) *AnalysisTask {
  return &AnalysisTask{
    AnsqContext: ansqContext,
  }
}

func (t *AnalysisTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = tasks.NewTradingsTask(t.AnsqContext)
  }
  return t.TradingsTask
}

func (t *AnalysisTask) Flush() {
  t.Tradings().Flush()
}
