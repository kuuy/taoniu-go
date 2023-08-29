package tradings

import (
  models "taoniu.local/cryptos/models/dydx"
)

type PendingInfo struct {
  Symbol   string
  Quantity float64
}

type MarketsRepository interface {
  Price(symbol string, side int) (float64, error)
  Get(symbol string) (models.Market, error)
}

type AccountRepository interface {
  Balance() (map[string]float64, error)
}

type PositionRepository interface {
  Get(symbol string) (models.Position, error)
}

type OrdersRepository interface {
  Status(orderID string) string
  Create(symbol string, side string, price float64, quantity float64) (string, error)
  Cancel(orderID string) error
  Lost(symbol string, side string, quantity float64, timestamp int64) string
  Flush(orderID string) error
}
