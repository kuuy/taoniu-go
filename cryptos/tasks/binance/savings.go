package binance

import (
	"context"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/savings"
	tasks "taoniu.local/cryptos/tasks/binance/savings"
)

type SavingsTask struct {
	Db           *gorm.DB
	Ctx          context.Context
	ProductsTask *tasks.ProductsTask
}

func (t *SavingsTask) Products() *tasks.ProductsTask {
	if t.ProductsTask == nil {
		t.ProductsTask = &tasks.ProductsTask{}
		t.ProductsTask.Repository = &repositories.ProductsRepository{
			Db:  t.Db,
			Ctx: t.Ctx,
		}
	}
	return t.ProductsTask
}
