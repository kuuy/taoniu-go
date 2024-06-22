package streams

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Orders struct{}

type OrdersUpdatePayload struct {
  Symbol  string
  OrderId int64
  Status  string
}

func (h *Orders) Update(symbol string, orderId int64, status string) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersUpdatePayload{symbol, orderId, status})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:orders:update", payload), nil
}
