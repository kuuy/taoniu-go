package futures

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
