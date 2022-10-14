package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/klines"
	tasks "taoniu.local/cryptos/tasks/binance/spot/klines"
)

type KlinesTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *KlinesTask) Daily() *tasks.DailyTask {
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
