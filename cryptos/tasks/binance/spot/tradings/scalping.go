package tradings

import repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"

type ScalpingTask struct {
	Repository *repositories.ScalpingRepository
}

func (t *ScalpingTask) Place() error {
	return t.Repository.Place()
}

func (t *ScalpingTask) Flush() error {
	return t.Repository.Flush()
}
