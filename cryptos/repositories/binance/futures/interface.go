package futures

type TradingsTriggersRepository interface {
  Scan() []string
}

type RankingResult struct {
  Total int
  Data  []string
}

type GameblingPlan struct {
  TakePrice    float64
  TakeQuantity float64
  TakeAmount   float64
}
