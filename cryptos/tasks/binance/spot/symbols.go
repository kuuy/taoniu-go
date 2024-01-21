package spot

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type SymbolsTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.SymbolsRepository
}

func NewSymbolsTask(ansqContext *common.AnsqClientContext) *SymbolsTask {
  return &SymbolsTask{
    AnsqContext: ansqContext,
    Repository: &repositories.SymbolsRepository{
      Db:  ansqContext.Db,
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
    },
  }
}

func (t *SymbolsTask) Flush() error {
  return t.Repository.Flush()
}

func (t *SymbolsTask) Count() error {
  return t.Repository.Count()
}
