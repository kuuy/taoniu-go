package analysis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis/daily"
)

type DailyTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *DailyTask) Margin() *tasks.MarginTask {
	return &tasks.MarginTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *DailyTask) Flush() error {
	t.Margin().Flush()
	return nil
}
