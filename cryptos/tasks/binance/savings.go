package binance

import (
  "taoniu.local/cryptos/common"
  tasks "taoniu.local/cryptos/tasks/binance/savings"
)

type SavingsTask struct {
  AnsqContext  *common.AnsqClientContext
  ProductsTask *tasks.ProductsTask
}

func NewSavingsTask(ansqContext *common.AnsqClientContext) *SavingsTask {
  return &SavingsTask{
    AnsqContext: ansqContext,
  }
}

func (t *SavingsTask) Products() *tasks.ProductsTask {
  if t.ProductsTask == nil {
    t.ProductsTask = tasks.NewProductsTask(t.AnsqContext)
  }
  return t.ProductsTask
}
