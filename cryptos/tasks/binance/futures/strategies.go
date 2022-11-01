package futures

import (
	"context"
	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
	tasks "taoniu.local/cryptos/tasks/binance/futures/strategies"
)

type StrategiesTask struct {
	Db        *gorm.DB
	Rdb       *redis.Client
	Ctx       context.Context
	DailyTask *tasks.DailyTask
}

func (t *StrategiesTask) Daily() *tasks.DailyTask {
	if t.DailyTask == nil {
		t.DailyTask = &tasks.DailyTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.DailyTask.Repository = &repositories.DailyRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.DailyTask
}
