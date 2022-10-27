package daily

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/daily/margin"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/daily/margin"
)

type MarginTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *MarginTask) Isolated() *tasks.IsolatedTask {
	return &tasks.IsolatedTask{
		Repository: &repositories.IsolatedRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *MarginTask) Flush() {
	t.Isolated().Flush()
}
