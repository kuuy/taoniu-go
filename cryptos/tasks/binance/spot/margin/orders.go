package margin

import (
  "strconv"
  "strings"
  "time"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
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
    },
  }
}

func (t *OrdersTask) Flush() error {
  orders, err := t.AnsqContext.Rdb.SMembers(t.AnsqContext.Ctx, "binance:spot:margin:orders:flush").Result()
  if err != nil {
    return nil
  }
  for _, order := range orders {
    data := strings.Split(order, ",")
    symbol := data[0]
    orderID, _ := strconv.ParseInt(data[1], 10, 64)
    isIsolated, _ := strconv.ParseBool(data[2])
    t.Repository.Flush(symbol, orderID, isIsolated)
  }
  return nil
}

func (t *OrdersTask) Fix() error {
  t.Repository.Fix(time.Now().Add(-30*time.Minute), 20)
  return nil
}
