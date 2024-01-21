package fishers

import (
  "taoniu.local/cryptos/common"
  savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
)

type GridsTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.GridsRepository
}

func NewGridsTask(ansqContext *common.AnsqClientContext) *GridsTask {
  return &GridsTask{
    AnsqContext: ansqContext,
    Repository: &repositories.GridsRepository{
      Db: ansqContext.Db,
      AccountRepository: &isolatedRepositories.AccountRepository{
        Db:  ansqContext.Db,
        Rdb: ansqContext.Rdb,
        Ctx: ansqContext.Ctx,
      },
      ProductsRepository: &savingsRepositories.ProductsRepository{
        Db: ansqContext.Db,
      },
    },
  }
}

func (t *GridsTask) Earn() error {
  t.Repository.Earn()
  return nil
}
