package patterns

import (
  futuresRepositories "taoniu.local/cryptos/repositories/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx/patterns"
)

type CandlesticksTask struct {
  Repository        *repositories.CandlesticksRepository
  MarketsRepository *futuresRepositories.MarketsRepository
}

func (t *CandlesticksTask) Clean() error {
  symbols := t.MarketsRepository.Symbols()
  for _, symbol := range symbols {
    t.Repository.Clean(symbol)
  }
  return nil
}
