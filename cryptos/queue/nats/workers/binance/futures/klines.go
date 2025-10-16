package futures

import (
  "encoding/json"
  "errors"
  "fmt"
  "gorm.io/gorm"
  "strconv"
  "time"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Klines struct {
  NatsContext *common.NatsContext
  Repository  *repositories.KlinesRepository
}

func NewKlines(natsContext *common.NatsContext) *Klines {
  h := &Klines{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.KlinesRepository{
    Db:  h.NatsContext.Db,
    Ctx: h.NatsContext.Ctx,
  }
  return h
}

func (h *Klines) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_KLINES_FLUSH, h.Flush)
  return nil
}

func (h *Klines) Flush(m *nats.Msg) {
  var payload *KlinesFlushPayload
  json.Unmarshal(m.Data, &payload)

  mutex := common.NewMutex(
    h.NatsContext.Rdb,
    h.NatsContext.Ctx,
    fmt.Sprintf(config.LOCKS_KLINES_FLUSH, payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  data, _ := h.NatsContext.Rdb.HMGet(
    h.NatsContext.Ctx,
    fmt.Sprintf(
      "binance:futures:klines:%s:%s:%v",
      payload.Interval,
      payload.Symbol,
      payload.Timestamp,
    ),
    "open",
    "close",
    "high",
    "low",
    "volume",
    "quota",
  ).Result()
  if len(data) == 0 {
    return
  }
  for i := 0; i < len(data); i++ {
    if data[i] == nil {
      return
    }
  }
  open, _ := strconv.ParseFloat(data[0].(string), 64)
  close, _ := strconv.ParseFloat(data[1].(string), 64)
  high, _ := strconv.ParseFloat(data[2].(string), 64)
  low, _ := strconv.ParseFloat(data[3].(string), 64)
  volume, _ := strconv.ParseFloat(data[4].(string), 64)
  quota, _ := strconv.ParseFloat(data[5].(string), 64)

  kline, err := h.Repository.Get(payload.Symbol, payload.Interval, payload.Timestamp)
  if errors.Is(err, gorm.ErrRecordNotFound) {
    h.Repository.Create(
      payload.Symbol,
      payload.Interval,
      open,
      close,
      high,
      low,
      volume,
      quota,
      payload.Timestamp,
    )
  } else {
    h.Repository.Updates(kline, map[string]interface{}{
      "open":   open,
      "close":  close,
      "high":   high,
      "low":    low,
      "volume": volume,
      "quota":  quota,
    })
  }

  message, _ := json.Marshal(&KlinesUpdatePayload{
    Symbol:   payload.Symbol,
    Interval: payload.Interval,
  })
  h.NatsContext.Conn.Publish(config.NATS_KLINES_UPDATE, message)
  h.NatsContext.Conn.Flush()
}
