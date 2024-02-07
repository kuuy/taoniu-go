package spot

import (
  "taoniu.local/cryptos/common"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Klines
  Repository         *repositories.KlinesRepository
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewKlinesTask(ansqContext *common.AnsqClientContext) *KlinesTask {
  return &KlinesTask{
    AnsqContext: ansqContext,
    Repository: &repositories.KlinesRepository{
      Db: ansqContext.Db,
    },
    SymbolsRepository: &repositories.SymbolsRepository{
      Db: ansqContext.Db,
    },
  }
}

func (t *KlinesTask) Clean() error {
  symbols := t.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    t.Repository.Clean(symbol)
  }
  return nil
}
