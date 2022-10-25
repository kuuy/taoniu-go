package plans

import repositories "taoniu.local/cryptos/repositories/binance/spot/plans"

type DailyTask struct {
	Repository *repositories.DailyRepository
}

func (t *DailyTask) Flush() error {
	return t.Repository.Flush()
}

func (t *DailyTask) Fix() error {
	return t.Repository.Fix(3600)
}
