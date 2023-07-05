package tradings

import (
  "gorm.io/datatypes"
  futuresModels "taoniu.local/cryptos/models/binance/futures"
)

type PendingInfo struct {
  Symbol   string
  Quantity float64
}

type SymbolsRepository interface {
  Price(symbol string) (float64, error)
  Get(symbol string) (futuresModels.Symbol, error)
  Filters(params datatypes.JSONMap) (float64, float64, error)
}

type PositionRepository interface {
  Get(symbol string, side int) (futuresModels.Position, error)
}

type OrdersRepository interface {
  Status(symbol string, orderID int64) string
  Create(symbol string, positionSide string, side string, price float64, quantity float64) (int64, error)
  Take(symbol string, positionSide string, price float64) (int64, error)
  Stop(symbol string, positionSide string, price float64) (int64, error)
  Cancel(symbol string, orderId int64) error
  Lost(symbol string, positionSide string, side string, price float64, timestamp int64) int64
  Flush(symbol string, orderID int64) error
}
