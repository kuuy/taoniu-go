package futures

import (
  "context"
  "fmt"
  "log"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type TriggersTask struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (t *TriggersTask) Flush() error {
  var triggers []*models.Trigger
  t.Db.Model(&models.Trigger{}).Where("status", 1).Find(&triggers)
  for _, entity := range triggers {
    data, _ := t.Rdb.HMGet(
      t.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:1d:%s:%s",
        entity.Symbol,
        time.Now().Format("0102"),
      ),
      "take_profit_price",
      "stop_loss_point",
    ).Result()
    if data[0] == nil || data[1] == nil {
      log.Println("indicators empty", entity.Symbol)
      continue
    }

    takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
    stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

    var price float64
    if entity.Side == 1 {
      price = stopPrice
    } else {
      price = takePrice
    }

    t.Db.Model(&entity).Update("price", price)
  }

  return nil
}
