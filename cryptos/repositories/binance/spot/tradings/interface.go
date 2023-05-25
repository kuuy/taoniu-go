package tradings

type PendingInfo struct {
  Symbol   string
  Quantity float64
}

type AnalysisRepository interface {
  Summary(exchange string, symbol string, interval string) (map[string]interface{}, error)
}

type SymbolsRepository interface {
  Price(symbol string) (float64, error)
  Adjust(symbol string, price float64, amount float64) (float64, float64, error)
}

type AccountRepository interface {
  Balance(symbol string) (float64, float64, error)
  Flush() error
}

type OrdersRepository interface {
  Status(symbol string, orderID int64) string
  Create(symbol string, side string, price float64, quantity float64) (int64, error)
  Lost(symbol string, side string, price float64, timestamp int64) int64
  Flush(symbol string, orderID int64) error
}

type SpotSymbolsRepository interface {
  Price(symbol string) (float64, error)
  Adjust(symbol string, price float64, amount float64) (float64, float64, error)
}
