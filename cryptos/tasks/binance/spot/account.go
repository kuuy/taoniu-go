package spot

import (
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type AccountTask struct {
	Repository *repositories.AccountRepository
}

func (t *AccountTask) Flush() error {
	return t.Repository.Flush()
}
