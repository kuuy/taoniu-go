package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/plans"
	tasks "taoniu.local/cryptos/tasks/binance/spot/plans"
)

type PlansTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *PlansTask) Daily() *tasks.DailyTask {
	return &tasks.DailyTask{
		Repository: &repositories.DailyRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}
