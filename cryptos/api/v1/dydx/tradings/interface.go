package tradings

type ScalpingTradingInfo struct {
  ID           string  `json:"id"`
  Symbol       string  `json:"symbol"`
  ScalpingID   string  `json:"scalping_id"`
  PlanID       string  `json:"plan_id"`
  BuyPrice     float64 `json:"buy_price"`
  SellPrice    float64 `json:"sell_price"`
  BuyQuantity  float64 `json:"buy_quantity"`
  SellQuantity float64 `json:"sell_quantity"`
  BuyOrderId   string  `json:"buy_order_id"`
  SellOrderId  string  `json:"sell_order_id"`
  Status       int     `json:"status"`
  CreatedAt    int64   `json:"created_at"`
  UpdatedAt    int64   `json:"updated_at"`
}

type TriggerTradingInfo struct {
  ID           string  `json:"id"`
  Symbol       string  `json:"symbol"`
  TriggerID    string  `json:"trigger_id"`
  BuyPrice     float64 `json:"buy_price"`
  SellPrice    float64 `json:"sell_price"`
  BuyQuantity  float64 `json:"buy_quantity"`
  SellQuantity float64 `json:"sell_quantity"`
  BuyOrderId   string  `json:"buy_order_id"`
  SellOrderId  string  `json:"sell_order_id"`
  Status       int     `json:"status"`
  CreatedAt    int64   `json:"created_at"`
  UpdatedAt    int64   `json:"updated_at"`
}
