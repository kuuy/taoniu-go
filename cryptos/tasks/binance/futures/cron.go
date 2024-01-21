package futures

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type CronTask struct {
  AnsqContext *common.AnsqClientContext
  SymbolsTask *SymbolsTask
}

func NewCronTask(ansqContext *common.AnsqClientContext) *CronTask {
  return &CronTask{
    AnsqContext: ansqContext,
  }
}

func (t *CronTask) Symbols() *SymbolsTask {
  if t.SymbolsTask == nil {
    t.SymbolsTask = &SymbolsTask{}
    t.SymbolsTask.Repository = &repositories.SymbolsRepository{
      Db:  t.AnsqContext.Db,
      Rdb: t.AnsqContext.Rdb,
      Ctx: t.AnsqContext.Ctx,
    }
  }
  return t.SymbolsTask
}

func (t *CronTask) Hourly() error {
  t.Symbols().Flush()
  t.Symbols().Count()
  return nil
}
