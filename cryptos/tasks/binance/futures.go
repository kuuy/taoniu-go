package binance

import (
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	repositories "taoniu.local/cryptos/repositories/binance/futures"
	tasks "taoniu.local/cryptos/tasks/binance/futures"
)

type FuturesTask struct {
	Db             *gorm.DB
	Rdb            *redis.Client
	Ctx            context.Context
	CronTask       *tasks.CronTask
	SymbolsTask    *tasks.SymbolsTask
	TickersTask    *tasks.TickersTask
	KlinesTask     *tasks.KlinesTask
	IndicatorsTask *tasks.IndicatorsTask
	StrategiesTask *tasks.StrategiesTask
	PlansTask      *tasks.PlansTask
}

func (t *FuturesTask) Cron() *tasks.CronTask {
	if t.CronTask == nil {
		t.CronTask = &tasks.CronTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.CronTask
}

func (t *FuturesTask) Symbols() *tasks.SymbolsTask {
	if t.SymbolsTask == nil {
		t.SymbolsTask = &tasks.SymbolsTask{}
		t.SymbolsTask.Repository = &repositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.SymbolsTask
}

func (t *FuturesTask) Tickers() *tasks.TickersTask {
	if t.TickersTask == nil {
		t.TickersTask = &tasks.TickersTask{
			Db: t.Db,
		}
		t.TickersTask.Repository = &repositories.TickersRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.TickersTask
}

func (t *FuturesTask) Klines() *tasks.KlinesTask {
	if t.KlinesTask == nil {
		t.KlinesTask = &tasks.KlinesTask{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.KlinesTask.Repository = &repositories.KlinesRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.KlinesTask
}

func (t *FuturesTask) Indicators() *tasks.IndicatorsTask {
	if t.IndicatorsTask == nil {
		t.IndicatorsTask = &tasks.IndicatorsTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.IndicatorsTask
}

func (t *FuturesTask) Strategies() *tasks.StrategiesTask {
	if t.StrategiesTask == nil {
		t.StrategiesTask = &tasks.StrategiesTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.StrategiesTask
}

func (t *FuturesTask) Plans() *tasks.PlansTask {
	if t.PlansTask == nil {
		t.PlansTask = &tasks.PlansTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.PlansTask
}

func (t *FuturesTask) Flush() {
	t.Indicators().Daily().Pivot()
	t.Indicators().Daily().Atr(14, 100)
	t.Plans().Daily().Flush()
}

func (t *FuturesTask) Clean() {
	t.Klines().Clean()
}

func (t *FuturesTask) Sync() {
}
