package dydx

type TradingsTriggersRepository interface {
  Scan() []string
}

type RankingResult struct {
  Total int
  Data  []string
}
