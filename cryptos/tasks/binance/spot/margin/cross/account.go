package cross

import (
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
)

type AccountTask struct {
  Repository *repositories.AccountRepository
}

func (t *AccountTask) Flush() error {
  return t.Repository.Flush()
}
