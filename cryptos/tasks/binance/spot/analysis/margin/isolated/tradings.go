package isolated

import (
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/margin/isolated/tradings"
)

type TradingsTask struct {
	Db *gorm.DB
}

func (t *TradingsTask) Fishers() *tasks.FishersTask {
	return &tasks.FishersTask{
		Db: t.Db,
	}
}

func (t *TradingsTask) Flush() error {
	t.Fishers().Flush()
	return nil
}
