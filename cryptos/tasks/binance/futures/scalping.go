package futures

import (
  "fmt"
  "log"
  "strconv"
  "time"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
)

type ScalpingTask struct {
  AnsqContext *common.AnsqClientContext
}

func NewScalpingTask(ansqContext *common.AnsqClientContext) *ScalpingTask {
  return &ScalpingTask{
    AnsqContext: ansqContext,
  }
}

func (t *ScalpingTask) Flush() error {
  var scalping []*models.Scalping
  t.AnsqContext.Db.Model(&models.Scalping{}).Where("status", 1).Find(&scalping)
  for _, entity := range scalping {
    data, _ := t.AnsqContext.Rdb.HMGet(
      t.AnsqContext.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:4h:%s:%s",
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

    t.AnsqContext.Db.Model(&entity).Update("price", price)
  }

  return nil
}
