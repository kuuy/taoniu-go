package patterns

import (
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/patterns"
)

type CandlesticksTask struct {
  Repository        *repositories.CandlesticksRepository
  SymbolsRepository *futuresRepositories.SymbolsRepository
}

func (t *CandlesticksTask) Clean() error {
  symbols := t.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    t.Repository.Clean(symbol)
  }
  return nil
}
