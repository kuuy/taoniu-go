package fishers

import (
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
)

type GridsTask struct {
	Repository *repositories.GridsRepository
}

func (t *GridsTask) Collect() error {
	t.Repository.Collect()
	return nil
}
