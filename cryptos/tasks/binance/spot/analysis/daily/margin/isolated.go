package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/daily/margin/isolated"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/daily/margin/isolated"
)

type IsolatedTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *IsolatedTask) Profits() *tasks.ProfitsTask {
	return &tasks.ProfitsTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Repository: &repositories.ProfitsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *IsolatedTask) Flush() {
	t.Profits().Flush()
}
