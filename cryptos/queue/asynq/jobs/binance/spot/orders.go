package spot

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Orders struct{}

type OrdersOpenPayload struct {
  Symbol string `json:"symbol"`
}

type OrdersFlushPayload struct {
  Symbol  string `json:"symbol"`
  OrderId int64  `json:"order_id"`
}

type OrdersSyncPayload struct {
  Symbol    string `json:"symbol"`
  StartTime int64  `json:"start_time"`
  limit     int    `json:"limit"`
}

func (h *Orders) Open(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersOpenPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_ORDERS_OPEN, payload), nil
}

func (h *Orders) Flush(symbol string, orderId int64) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersFlushPayload{symbol, orderId})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_ORDERS_FLUSH, payload), nil
}

func (h *Orders) Sync(symbol string, startTime int64, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersSyncPayload{symbol, startTime, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_ORDERS_SYNC, payload), nil
}
