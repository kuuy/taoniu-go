package tradings

import (
  "gorm.io/datatypes"
  spotModels "taoniu.local/cryptos/models/binance/spot"
)

type SymbolsRepository interface {
  Price(symbol string) (float64, error)
  Get(symbol string) (spotModels.Symbol, error)
  Filters(params datatypes.JSONMap) (float64, float64, float64, error)
}

type AccountRepository interface {
  Balance(symbol string) (map[string]float64, error)
}

type MarginAccountRepository interface {
  Loan(asset string, symbol string, amount float64, isIsolated bool) (int64, error)
}

type OrdersRepository interface {
  Status(symbol string, orderId int64) string
  Create(symbol string, side string, price float64, quantity float64, isIsolated bool) (int64, error)
  Lost(symbol string, side string, price float64, timestamp int64) int64
  Flush(symbol string, orderId int64, isIsolated bool) error
}
