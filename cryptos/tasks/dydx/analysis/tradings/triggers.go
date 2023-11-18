package tradings

import (
  repositories "taoniu.local/cryptos/repositories/dydx/analysis/tradings"
)

type TriggersTask struct {
  Repository *repositories.TriggersRepository
}

func (t *TriggersTask) Flush() error {
  t.Repository.Flush(1)
  t.Repository.Flush(2)
  return nil
}
