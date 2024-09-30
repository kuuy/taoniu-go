package spot

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
