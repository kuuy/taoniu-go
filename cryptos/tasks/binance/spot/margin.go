package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin"
)

type MarginTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *MarginTask) Isolated() *tasks.IsolatedTask {
	return &tasks.IsolatedTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}
