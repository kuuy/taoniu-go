package streams

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Orders struct{}

type OrdersUpdatePayload struct {
  Symbol  string
  OrderID int64
  Status  string
}

func (h *Orders) Update(symbol string, orderID int64, status string) (*asynq.Task, error) {
  payload, err := json.Marshal(OrdersUpdatePayload{symbol, orderID, status})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:orders:update", payload), nil
}
