package isolated

import (
  "taoniu.local/cryptos/common"
  marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type OrdersTask struct {
  AnsqContext *common.AnsqClientContext
  Repository  *repositories.OrdersRepository
}

func NewOrdersTask(ansqContext *common.AnsqClientContext) *OrdersTask {
  return &OrdersTask{
    AnsqContext: ansqContext,
    Repository: &repositories.OrdersRepository{
      Db:  ansqContext.Db,
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
      Parent: &marginRepositories.OrdersRepository{
        Db:  ansqContext.Db,
        Rdb: ansqContext.Rdb,
        Ctx: ansqContext.Ctx,
      },
    },
  }
}

func (t *OrdersTask) Open() error {
  symbols, _ := t.AnsqContext.Rdb.SMembers(t.AnsqContext.Ctx, "binance:spot:websocket:symbols").Result()
  for _, symbol := range symbols {
    t.Repository.Open(symbol)
  }
  return nil
}

func (t *OrdersTask) Sync() error {
  symbols, _ := t.AnsqContext.Rdb.SMembers(t.AnsqContext.Ctx, "binance:spot:margin:isolated:symbols").Result()
  for _, symbol := range symbols {
    t.Repository.Sync(symbol, 20)
  }
  return nil
}
