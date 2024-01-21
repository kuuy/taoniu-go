package spot

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type GridsTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.GridsRepository
}

func NewGridsTask(ansqContext *common.AnsqClientContext) *GridsTask {
  return &GridsTask{
    AnsqContext: ansqContext,
    Repository: &repositories.GridsRepository{
      Db:  ansqContext.Db,
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
    },
  }
}

func (t *GridsTask) Flush() error {
  symbols, _ := t.AnsqContext.Rdb.SMembers(t.AnsqContext.Ctx, "binance:spot:grids:symbols").Result()
  for _, symbol := range symbols {
    t.Repository.Flush(symbol)
  }
  return nil
}
