package spot

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Orders struct{}

type OrdersOpenPayload struct {
  Symbol string `json:"symbol"`
}

type OrdersFlushPayload struct {
  Symbol  string `json:"symbol"`
  OrderID int64  `json:"order_id"`
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
  return asynq.NewTask("binance:spot:orders:open", payload), nil
}

func (h *Orders) Flush(symbol string, orderID int64) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersFlushPayload{symbol, orderID})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:orders:flush", payload), nil
}

func (h *Orders) Sync(symbol string, startTime int64, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersSyncPayload{symbol, startTime, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:orders:sync", payload), nil
}
