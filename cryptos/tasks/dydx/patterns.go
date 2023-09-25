package dydx

import (
  "gorm.io/gorm"

  repositories "taoniu.local/cryptos/repositories/dydx"
  patternsRepositories "taoniu.local/cryptos/repositories/dydx/patterns"
  tasks "taoniu.local/cryptos/tasks/dydx/patterns"
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
    t.CandlesticksTask.MarketsRepository = &repositories.MarketsRepository{
      Db: t.Db,
    }
  }
  return t.CandlesticksTask
}
