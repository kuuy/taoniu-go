package dydx

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

type ScalpingTask struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (t *ScalpingTask) Flush() error {
  var scalping []*models.Scalping
  t.Db.Model(&models.Scalping{}).Where("status", 1).Find(&scalping)
  for _, entity := range scalping {
    data, _ := t.Rdb.HMGet(
      t.Ctx,
      fmt.Sprintf(
        "dydx:indicators:4h:%s:%s",
        entity.Symbol,
        time.Now().Format("0102"),
      ),
      "vah",
      "val",
    ).Result()
    if len(data) == 0 || data[0] == nil || data[1] == nil {
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
