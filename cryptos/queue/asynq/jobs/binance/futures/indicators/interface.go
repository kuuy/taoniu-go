package indicators

type IndicatorPayload struct {
  Symbol string
  Period int
  Limit  int
}

type PivotPayload struct {
  Symbol string
}

type KdjPayload struct {
  Symbol      string
  LongPeriod  int
  ShortPeriod int
  Limit       int
}

type VolumeProfilePayload struct {
  Symbol string
  Limit  int
}
