package analysis

import (
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/margin"
)

type MarginTask struct {
	Db *gorm.DB
}

func (t *MarginTask) Isolated() *tasks.IsolatedTask {
	return &tasks.IsolatedTask{
		Db: t.Db,
	}
}

func (t *MarginTask) Flush() error {
	t.Isolated().Flush()
	return nil
}
