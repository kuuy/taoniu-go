package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis"
)

type AnalysisTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *AnalysisTask) Margin() *tasks.MarginTask {
	return &tasks.MarginTask{
		Db: t.Db,
	}
}
