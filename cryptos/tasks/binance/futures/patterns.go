package futures

import (
  "gorm.io/gorm"

  repositories "taoniu.local/cryptos/repositories/binance/futures"
  patternsRepositories "taoniu.local/cryptos/repositories/binance/futures/patterns"
  tasks "taoniu.local/cryptos/tasks/binance/futures/patterns"
)

type PatternsTask struct {
  Db               *gorm.DB
  CandlesticksTask *tasks.CandlesticksTask
}

func (t *PatternsTask) Candlesticks() *tasks.CandlesticksTask {
  if t.CandlesticksTask == nil {
    t.CandlesticksTask = &tasks.CandlesticksTask{}
    t.CandlesticksTask.Repository = &patternsRepositories.CandlesticksRepository{
      Db: t.Db,
    }
    t.CandlesticksTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
  }
  return t.CandlesticksTask
}
