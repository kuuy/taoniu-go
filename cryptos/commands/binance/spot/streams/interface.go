package streams

type KlinesFlushPayload struct {
  Symbol    string `json:"symbol"`
  Interval  string `json:"interval"`
  Timestamp int64  `json:"timestamp"`
}
