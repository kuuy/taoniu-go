package futures

type AccountUpdatePayload struct {
  Symbol           string  `json:"symbol"`
  Balance          float64 `json:"balance"`
  Free             float64 `json:"free"`
  UnrealizedProfit float64 `json:"unrealized_profit"`
  Margin           float64 `json:"margin"`
  InitialMargin    float64 `json:"initial_margin"`
  MaintMargin      float64 `json:"maint_margin"`
}

type OrdersUpdatePayload struct {
  Symbol  string `json:"symbol"`
  OrderID int64  `json:"order_id"`
  Status  string `json:"status"`
}

type TickersUpdatePayload struct {
  Symbol    string  `json:"symbol"`
  Open      float64 `json:"open"`
  Price     float64 `json:"price"`
  High      float64 `json:"high"`
  Low       float64 `json:"low"`
  Volume    float64 `json:"volume"`
  Quota     float64 `json:"quota"`
  Timestamp int64   `json:"timestamp"`
}
