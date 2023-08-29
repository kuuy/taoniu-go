package dydx

import repositories "taoniu.local/cryptos/repositories/dydx"

type MarketsTask struct {
  Repository *repositories.MarketsRepository
}

func (t *MarketsTask) Flush() error {
  return t.Repository.Flush()
}
