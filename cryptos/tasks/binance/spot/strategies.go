package spot

import (
	"context"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	repositories "taoniu.local/cryptos/repositories/binance/spot/strategies"
	tasks "taoniu.local/cryptos/tasks/binance/spot/strategies"
)

type StrategiesTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *StrategiesTask) Daily() *tasks.DailyTask {
	return &tasks.DailyTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Repository: &repositories.DailyRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}
