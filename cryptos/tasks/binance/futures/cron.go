package futures

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type CronTask struct {
	Db          *gorm.DB
	Rdb         *redis.Client
	Ctx         context.Context
	SymbolsTask *SymbolsTask
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

func (t *CronTask) Hourly() error {
	t.Symbols().Flush()
	t.Symbols().Count()
	return nil
}
