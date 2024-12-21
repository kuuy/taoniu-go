package gambling

import (
  "gorm.io/datatypes"
  models "taoniu.local/cryptos/models/binance/futures"
)

type PendingInfo struct {
  Symbol   string
  Quantity float64
}

type SymbolsRepository interface {
  Price(symbol string) (float64, error)
  Get(symbol string) (*models.Symbol, error)
  Filters(params datatypes.JSONMap) (float64, float64, float64, error)
}

type AccountRepository interface {
  Balance(symbol string) (map[string]float64, error)
}

type PositionRepository interface {
  Get(symbol string, side int) (*models.Position, error)
  Capital(capital float64, entryAmount float64, place int) (float64, error)
  Ratio(capital float64, entryAmount float64) float64
  BuyQuantity(side int, buyAmount float64, entryPrice float64, entryAmount float64) float64
  SellPrice(side int, entryPrice float64, entryAmount float64) float64
  TakePrice(entryPrice float64, side int, tickSize float64) float64
  StopPrice(maxCapital float64, side int, price float64, leverage int, entryPrice float64, entryQuantity float64, tickSize float64, stepSize float64) (float64, error)
}

type OrdersRepository interface {
  Status(symbol string, orderId int64) string
  Create(symbol string, positionSide string, side string, price float64, quantity float64) (int64, error)
  Take(symbol string, positionSide string, price float64) (int64, error)
  Stop(symbol string, positionSide string, price float64) (int64, error)
  Cancel(symbol string, orderId int64) error
  Lost(symbol string, positionSide string, side string, quantity float64, timestamp int64) int64
  Flush(symbol string, orderId int64) error
}
