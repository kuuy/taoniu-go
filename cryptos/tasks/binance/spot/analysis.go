package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis"
	"taoniu.local/cryptos/tasks/binance/spot/analysis/daily"
)

type AnalysisTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *AnalysisTask) Margin() *daily.MarginTask {
	return &daily.MarginTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *AnalysisTask) Daily() *tasks.DailyTask {
	return &tasks.DailyTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}
