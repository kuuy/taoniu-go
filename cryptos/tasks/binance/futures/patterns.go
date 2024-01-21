package futures

import (
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/futures"
  patternsRepositories "taoniu.local/cryptos/repositories/binance/futures/patterns"
  tasks "taoniu.local/cryptos/tasks/binance/futures/patterns"
)

type PatternsTask struct {
  AnsqContext      *common.AnsqClientContext
  CandlesticksTask *tasks.CandlesticksTask
}

func NewPatternsTask(ansqContext *common.AnsqClientContext) *PatternsTask {
  return &PatternsTask{
    AnsqContext: ansqContext,
  }
}

func (t *PatternsTask) Candlesticks() *tasks.CandlesticksTask {
  if t.CandlesticksTask == nil {
    t.CandlesticksTask = &tasks.CandlesticksTask{}
    t.CandlesticksTask.Repository = &patternsRepositories.CandlesticksRepository{
      Db: t.AnsqContext.Db,
    }
    t.CandlesticksTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.AnsqContext.Db,
    }
  }
  return t.CandlesticksTask
}
