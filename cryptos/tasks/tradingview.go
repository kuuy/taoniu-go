package tasks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/tradingview"
	tasks "taoniu.local/cryptos/tasks/tradingview"
)

type TradingviewTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	AnalysisTask *tasks.AnalysisTask
}

func (t *TradingviewTask) Analysis() *tasks.AnalysisTask {
	if t.AnalysisTask == nil {
		t.AnalysisTask = &tasks.AnalysisTask{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.AnalysisTask.Repository = &repositories.AnalysisRepository{
			Db: t.Db,
		}
		t.AnalysisTask.SymbolRepository = &spotRepositories.SymbolsRepository{
			Db: t.Db,
		}
	}
	return t.AnalysisTask
}
