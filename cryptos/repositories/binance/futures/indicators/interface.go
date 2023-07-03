package indicators

type RankingScore struct {
  Symbol string
  Value  float64
  Data   []string
}

type RankingResult struct {
  Total int
  Data  []string
}
