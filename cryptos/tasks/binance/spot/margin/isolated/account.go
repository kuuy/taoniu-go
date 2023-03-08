package isolated

import (
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type AccountTask struct {
	Repository *repositories.AccountRepository
}

func (t *AccountTask) Flush() error {
	return t.Repository.Flush()
}

func (t *AccountTask) Collect() error {
	return t.Repository.Collect()
}
