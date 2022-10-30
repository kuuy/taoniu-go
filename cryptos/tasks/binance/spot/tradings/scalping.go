package tradings

import repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"

type ScalpingTask struct {
	Repository *repositories.ScalpingRepository
}

func (t *ScalpingTask) Flush() error {
	return t.Repository.Flush()
}

func (t *ScalpingTask) Update() error {
	return t.Repository.Update()
}
