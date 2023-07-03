package indicators

type IndicatorPayload struct {
  Symbol      string
  Period      int
  LongPeriod  int
  ShortPeriod int
  Limit       int
}
