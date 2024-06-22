package spot

type AccountUpdatePayload struct {
  Asset string  `json:"asset"`
  Free  float64 `json:"free"`
  Lock  float64 `json:"lock"`
}

type OrdersUpdatePayload struct {
  Symbol  string `json:"symbol"`
  OrderId int64  `json:"order_id"`
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
