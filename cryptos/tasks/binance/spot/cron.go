package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type CronTask struct {
	Db          *gorm.DB
	Rdb         *redis.Client
	Ctx         context.Context
	SymbolsTask *SymbolsTask
	GridsTask   *GridsTask
	MarginTask  *MarginTask
}

func (t *CronTask) Symbols() *SymbolsTask {
	if t.SymbolsTask == nil {
		t.SymbolsTask = &SymbolsTask{}
		t.SymbolsTask.Repository = &repositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.SymbolsTask
}

func (t *CronTask) Grids() *GridsTask {
	if t.GridsTask == nil {
		t.GridsTask = &GridsTask{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.GridsTask.Repository = &repositories.GridsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.GridsTask
}

func (t *CronTask) Margin() *MarginTask {
	if t.MarginTask == nil {
		t.MarginTask = &MarginTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.MarginTask
}

func (t *CronTask) Hourly() error {
	t.Symbols().Flush()
	t.Symbols().Count()
	t.Grids().Flush()
	t.Margin().Sync()

	return nil
}
