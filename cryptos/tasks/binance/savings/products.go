package savings

import (
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/savings"
)

type ProductsTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.ProductsRepository
}

func NewProductsTask(ansqContext *common.AnsqClientContext) *ProductsTask {
  return &ProductsTask{
    AnsqContext: ansqContext,
    Repository: &repositories.ProductsRepository{
      Db:  ansqContext.Db,
      Ctx: ansqContext.Ctx,
    },
  }
}

func (t *ProductsTask) Flush() error {
  return t.Repository.Flush()
}
