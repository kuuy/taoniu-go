package tradings

import (
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings/fishers"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/tradings/fishers"
)

type FishersTask struct {
	Db *gorm.DB
}

func (t *FishersTask) Grids() *tasks.GridsTask {
	return &tasks.GridsTask{
		Repository: &repositories.GridsRepository{
			Db: t.Db,
		},
	}
}

func (t *FishersTask) Flush() error {
	t.Grids().Flush()
	return nil
}
