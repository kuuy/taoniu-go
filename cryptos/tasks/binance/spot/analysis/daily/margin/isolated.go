package margin

import (
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/daily/margin"
)

type IsolatedTask struct {
	Repository *repositories.IsolatedRepository
}

func (t *IsolatedTask) Flush() error {
	t.Repository.Grids()
	return nil
}
