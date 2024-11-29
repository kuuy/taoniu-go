package spot

type TickersFlushPayload struct {
  Symbols  []string
  UseProxy bool
}

type TickersUpdatePayload struct {
  Symbol    string  `json:"symbol"`
  Price     float64 `json:"price"`
  Open      float64 `json:"open"`
  High      float64 `json:"high"`
  Low       float64 `json:"low"`
  Volume    float64 `json:"volume"`
  Quota     float64 `json:"quota"`
  Timestamp int64   `json:"timestamp"`
}

type KlinesFlushPayload struct {
  Symbol   string
  Interval string
  Limit    int
  UseProxy bool
}

type KlinesUpdatePayload struct {
  Symbol    string  `json:"symbol"`
  Interval  string  `json:"interval"`
  Open      float64 `json:"open"`
  Close     float64 `json:"close"`
  High      float64 `json:"high"`
  Low       float64 `json:"low"`
  Volume    float64 `json:"volume"`
  Quota     float64 `json:"quota"`
  Timestamp int64   `json:"timestamp"`
}

type KlinesCleanPayload struct {
  Symbol string
}

type IndicatorPayload struct {
  Symbol   string
  Interval string
  Period   int
  Limit    int
}

type PivotPayload struct {
  Symbol   string
  Interval string
}

type KdjPayload struct {
  Symbol      string
  Interval    string
  LongPeriod  int
  ShortPeriod int
  Limit       int
}

type VolumeProfilePayload struct {
  Symbol   string
  Interval string
  Limit    int
}

type AndeanOscillatorPayload struct {
  Symbol   string
  Interval string
  Period   int
  Length   int
  Limit    int
}

type StrategyPayload struct {
  Symbol   string
  Interval string
}

type PlansPayload struct {
  Interval string
}

type OrdersOpenPayload struct {
  Symbol string `json:"symbol"`
}

type OrdersFlushPayload struct {
  Symbol  string `json:"symbol"`
  OrderId int64  `json:"order_id"`
}

type OrdersSyncPayload struct {
  Symbol    string `json:"symbol"`
  StartTime int64  `json:"start_time"`
  limit     int    `json:"limit"`
}
