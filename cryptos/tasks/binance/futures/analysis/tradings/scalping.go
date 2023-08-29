package tradings

import (
  repositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
)

type ScalpingTask struct {
  Repository *repositories.ScalpingRepository
}

func (t *ScalpingTask) Flush() error {
  t.Repository.Flush(1)
  t.Repository.Flush(2)
  return nil
}
