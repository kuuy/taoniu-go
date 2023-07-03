package spot

import (
  "context"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type OrdersTask struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.OrdersRepository
}

func (t *OrdersTask) Flush() error {
  orders, err := t.Rdb.SMembers(t.Ctx, "binance:spot:orders:flush").Result()
  if err != nil {
    return nil
  }
  for _, order := range orders {
    data := strings.Split(order, ",")
    symbol := data[0]
    orderID, _ := strconv.ParseInt(data[1], 10, 64)
    t.Repository.Flush(symbol, orderID)
  }
  return nil
}

func (t *OrdersTask) Open() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    t.Repository.Open(symbol)
  }
  return nil
}

func (t *OrdersTask) Sync() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    t.Repository.Sync(symbol, 20)
  }
  return nil
}

func (t *OrdersTask) Fix() error {
  t.Repository.Fix(time.Now().Add(-30*time.Minute), 20)
  return nil
}
