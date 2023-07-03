package fishers

import (
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
)

type GridsTask struct {
  Repository *repositories.GridsRepository
}

func (t *GridsTask) Earn() error {
  t.Repository.Earn()
  return nil
}
