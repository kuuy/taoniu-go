package plans

import repositories "taoniu.local/cryptos/repositories/binance/spot/plans"

type DailyTask struct {
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Flush() error {
	return t.Repository.Flush()
}
