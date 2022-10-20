package spot

import (
	"context"
	"github.com/gammazero/workerpool"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
	tasks "taoniu.local/cryptos/tasks/binance/spot/indicators"
)

type IndicatorsTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
	Wp  *workerpool.WorkerPool
}

func (t *IndicatorsTask) Daily() *tasks.DailyTask {
	return &tasks.DailyTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Wp:  t.Wp,
		Repository: &repositories.DailyRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}
