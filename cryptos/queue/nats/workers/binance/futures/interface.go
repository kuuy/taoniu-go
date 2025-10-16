package futures

type KlinesFlushPayload struct {
  Symbol    string `json:"symbol"`
  Interval  string `json:"interval"`
  Timestamp int64  `json:"timestamp"`
}

type KlinesUpdatePayload struct {
  Symbol   string `json:"symbol"`
  Interval string `json:"interval"`
}

type IndicatorsUpdatePayload struct {
  Symbol   string `json:"symbol"`
  Interval string `json:"interval"`
}

type StrategiesUpdatePayload struct {
  Symbol   string `json:"symbol"`
  Interval string `json:"interval"`
}

type PlansUpdatePayload struct {
  ID     string  `json:"id"`
  Side   int     `json:"side"`
  Amount float64 `json:"amount"`
}

type AccountUpdatePayload struct {
  Asset            string  `json:"asset"`
  Balance          float64 `json:"balance"`
  Free             float64 `json:"free"`
  UnrealizedProfit float64 `json:"unrealized_profit"`
  Margin           float64 `json:"margin"`
  InitialMargin    float64 `json:"initial_margin"`
  MaintMargin      float64 `json:"maint_margin"`
}
