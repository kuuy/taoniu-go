package spot

import repositories "taoniu.local/cryptos/repositories/binance/spot"

type TradingsTask struct {
	Repository *repositories.TradingsRepository
}

func (t *TradingsTask) Scalping() error {
	return t.Repository.Scalping()
}

func (t *TradingsTask) UpdateScalping() error {
	return t.Repository.UpdateScalping()
}
