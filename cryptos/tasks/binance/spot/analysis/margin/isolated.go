package margin

import (
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/margin/isolated"
)

type IsolatedTask struct {
	Db *gorm.DB
}

func (t *IsolatedTask) Tradings() *tasks.TradingsTask {
	return &tasks.TradingsTask{
		Db: t.Db,
	}
}

func (t *IsolatedTask) Flush() error {
	t.Tradings().Flush()
	return nil
}
