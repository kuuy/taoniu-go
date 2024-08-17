package tradings

type ScalpingInfo struct {
  ID          string  `json:"id"`
  Day         string  `json:"day"`
  BuysCount   int     `json:"buys_count"`
  SellsCount  int     `json:"sells_count"`
  BuysAmount  float64 `json:"buys_amount"`
  SellsAmount float64 `json:"sells_amount"`
}
