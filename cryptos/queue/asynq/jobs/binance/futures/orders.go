package futures

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Orders struct{}

type OrdersOpenPayload struct {
  Symbol string
}

type OrdersFlushPayload struct {
  Symbol  string
  OrderID int64
}

func (h *Orders) Open(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersOpenPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:orders:open", payload), nil
}

func (h *Orders) Flush(symbol string, orderID int64) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersFlushPayload{symbol, orderID})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:orders:flush", payload), nil
}
