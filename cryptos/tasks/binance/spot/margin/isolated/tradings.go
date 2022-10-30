package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated/tradings"
)

type TradingsTask struct {
	Db        *gorm.DB
	Rdb       *redis.Client
	Ctx       context.Context
	GridsTask *tasks.GridsTask
}

func (t *TradingsTask) Grids() *tasks.GridsTask {
	if t.GridsTask == nil {
		t.GridsTask = &tasks.GridsTask{
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
