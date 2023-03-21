package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	tasks "taoniu.local/cryptos/tasks/binance/spot/analysis"
)

type AnalysisTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	TradingsTask *tasks.TradingsTask
	MarginTask   *tasks.MarginTask
}

func (t *AnalysisTask) Tradings() *tasks.TradingsTask {
	if t.TradingsTask == nil {
		t.TradingsTask = &tasks.TradingsTask{
			Db: t.Db,
		}
	}
	return t.TradingsTask
}

func (t *AnalysisTask) Margin() *tasks.MarginTask {
	if t.MarginTask == nil {
		t.MarginTask = &tasks.MarginTask{
			Db: t.Db,
		}
	}
	return t.MarginTask
}

func (t *AnalysisTask) Flush() {
	t.Tradings().Flush()
	t.Margin().Flush()
}
