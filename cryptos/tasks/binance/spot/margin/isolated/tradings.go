package isolated

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated/tradings"
)

type TradingsTask struct {
  AnsqContext *common.AnsqClientContext
  FishersTask *tasks.FishersTask
  GridsTask   *tasks.GridsTask
}

func NewTradingsTask(ansqContext *common.AnsqClientContext) *TradingsTask {
  return &TradingsTask{
    AnsqContext: ansqContext,
  }
}

func (t *TradingsTask) Fishers() *tasks.FishersTask {
  if t.FishersTask == nil {
    t.FishersTask = tasks.NewFishersTask(t.AnsqContext)
  }
  return t.FishersTask
}

func (t *TradingsTask) Grids() *tasks.GridsTask {
  if t.GridsTask == nil {
    t.GridsTask = tasks.NewGridsTask(t.AnsqContext)
  }
  return t.GridsTask
}
