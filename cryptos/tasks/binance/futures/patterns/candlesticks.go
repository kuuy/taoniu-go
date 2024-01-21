package patterns

import (
  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/patterns"
)

type CandlesticksTask struct {
  AnsqContext       *common.AnsqClientContext
  Repository        *repositories.CandlesticksRepository
  SymbolsRepository *futuresRepositories.SymbolsRepository
}

func NewCandlesticksTask(ansqContext *common.AnsqClientContext) *CandlesticksTask {
  return &CandlesticksTask{
    AnsqContext: ansqContext,
    Repository: &repositories.CandlesticksRepository{
      Db: ansqContext.Db,
    },
    SymbolsRepository: &futuresRepositories.SymbolsRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *CandlesticksTask) Clean() error {
  symbols := t.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    t.Repository.Clean(symbol)
  }
  return nil
}
