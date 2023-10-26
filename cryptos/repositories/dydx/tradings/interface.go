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
  Capital(capital float64, entryAmount float64, place int) (float64, error)
  Ratio(capital float64, entryAmount float64) float64
  BuyQuantity(side int, buyAmount float64, entryPrice float64, entryAmount float64) float64
  SellPrice(side int, entryPrice float64, entryAmount float64) float64
  TakePrice(entryPrice float64, side int, tickSize float64) float64
  StopPrice(maxCapital float64, side int, price float64, leverage int, entryPrice float64, entryQuantity float64, tickSize float64, stepSize float64) (float64, error)
}

type OrdersRepository interface {
  Status(orderID string) string
  Create(symbol string, side string, price float64, quantity float64, positionSide string) (string, error)
  Cancel(orderID string) error
  Lost(symbol string, side string, quantity float64, timestamp int64) string
  Flush(orderID string) error
}
