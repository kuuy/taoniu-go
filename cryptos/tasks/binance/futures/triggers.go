package futures

import (
  "fmt"
  "log"
  "strconv"
  "time"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
)

type TriggersTask struct {
  AnsqContext *common.AnsqClientContext
}

func NewTriggersTask(ansqContext *common.AnsqClientContext) *TriggersTask {
  return &TriggersTask{
    AnsqContext: ansqContext,
  }
}

func (t *TriggersTask) Flush() error {
  var triggers []*models.Trigger
  t.AnsqContext.Db.Model(&models.Trigger{}).Where("status", 1).Find(&triggers)
  for _, entity := range triggers {
    data, _ := t.AnsqContext.Rdb.HMGet(
      t.AnsqContext.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:1d:%s:%s",
        entity.Symbol,
        time.Now().Format("0102"),
      ),
      "take_profit_price",
      "stop_loss_point",
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
