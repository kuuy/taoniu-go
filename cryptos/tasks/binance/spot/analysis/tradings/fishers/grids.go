package fishers

import repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings/fishers"

type GridsTask struct {
	Repository *repositories.GridsRepository
}

func (t *GridsTask) Flush() error {
	t.Repository.Flush()
	return nil
}