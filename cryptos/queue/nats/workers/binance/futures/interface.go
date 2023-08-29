package futures

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
  Amount float64 `json:"amount"`
}
