package binance

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	jobs "taoniu.local/cryptos/queue/jobs/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
	tasks "taoniu.local/cryptos/tasks/binance/spot"
)

type SpotTask struct {
	Db             *gorm.DB
	Rdb            *redis.Client
	Ctx            context.Context
	Asynq          *asynq.Client
	CronTask       *tasks.CronTask
	SymbolsTask    *tasks.SymbolsTask
	TickersTask    *tasks.TickersTask
	DepthTask      *tasks.DepthTask
	KlinesTask     *tasks.KlinesTask
	IndicatorsTask *tasks.IndicatorsTask
	StrategiesTask *tasks.StrategiesTask
	PlansTask      *tasks.PlansTask
	TradingsTask   *tasks.TradingsTask
	AccountTask    *tasks.AccountTask
	OrdersTask     *tasks.OrdersTask
	GridsTask      *tasks.GridsTask
	AnalysisTask   *tasks.AnalysisTask
	MarginTask     *tasks.MarginTask
}

func (t *SpotTask) Cron() *tasks.CronTask {
	if t.CronTask == nil {
		t.CronTask = &tasks.CronTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.CronTask
}

func (t *SpotTask) Symbols() *tasks.SymbolsTask {
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

func (t *SpotTask) Tickers() *tasks.TickersTask {
	if t.TickersTask == nil {
		t.TickersTask = &tasks.TickersTask{
			Rdb:   t.Rdb,
			Ctx:   t.Ctx,
			Asynq: t.Asynq,
		}
		t.TickersTask.Job = &jobs.Tickers{}
		t.TickersTask.Repository = &repositories.TickersRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.TickersTask.SymbolsRepository = &repositories.SymbolsRepository{
			Db: t.Db,
		}
	}
	return t.TickersTask
}

func (t *SpotTask) Depth() *tasks.DepthTask {
	if t.DepthTask == nil {
		t.DepthTask = &tasks.DepthTask{
			Asynq: t.Asynq,
		}
		t.DepthTask.Job = &jobs.Depth{}
		t.DepthTask.Repository = &repositories.DepthRepository{
			Db: t.Db,
		}
		t.DepthTask.SymbolsRepository = &repositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.DepthTask
}

func (t *SpotTask) Klines() *tasks.KlinesTask {
	if t.KlinesTask == nil {
		t.KlinesTask = &tasks.KlinesTask{
			Asynq: t.Asynq,
		}
		t.KlinesTask.Job = &jobs.Klines{}
		t.KlinesTask.Repository = &repositories.KlinesRepository{
			Db: t.Db,
		}
		t.KlinesTask.SymbolsRepository = &repositories.SymbolsRepository{
			Db: t.Db,
		}
	}
	return t.KlinesTask
}

func (t *SpotTask) Indicators() *tasks.IndicatorsTask {
	if t.IndicatorsTask == nil {
		t.IndicatorsTask = &tasks.IndicatorsTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.IndicatorsTask
}

func (t *SpotTask) Strategies() *tasks.StrategiesTask {
	if t.StrategiesTask == nil {
		t.StrategiesTask = &tasks.StrategiesTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.StrategiesTask
}

func (t *SpotTask) Plans() *tasks.PlansTask {
	if t.PlansTask == nil {
		t.PlansTask = &tasks.PlansTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.PlansTask
}

func (t *SpotTask) Tradings() *tasks.TradingsTask {
	if t.TradingsTask == nil {
		t.TradingsTask = &tasks.TradingsTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.TradingsTask
}

func (t *SpotTask) Account() *tasks.AccountTask {
	if t.AccountTask == nil {
		t.AccountTask = &tasks.AccountTask{}
		t.AccountTask.Repository = &repositories.AccountRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.AccountTask
}

func (t *SpotTask) Orders() *tasks.OrdersTask {
	if t.OrdersTask == nil {
		t.OrdersTask = &tasks.OrdersTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.OrdersTask.Repository = &repositories.OrdersRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.OrdersTask
}

func (t *SpotTask) Grids() *tasks.GridsTask {
	if t.GridsTask == nil {
		t.GridsTask = &tasks.GridsTask{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.GridsTask.Repository = &repositories.GridsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.GridsTask
}

func (t *SpotTask) Analysis() *tasks.AnalysisTask {
	if t.AnalysisTask == nil {
		t.AnalysisTask = &tasks.AnalysisTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.AnalysisTask
}

func (t *SpotTask) Margin() *tasks.MarginTask {
	if t.MarginTask == nil {
		t.MarginTask = &tasks.MarginTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.MarginTask
}

func (t *SpotTask) Flush() {
	t.Account().Flush()
	//t.Orders().Open()
	//t.Orders().Gets()
	t.Margin().Flush()
	t.Indicators().Daily().Pivot()
	t.Indicators().Daily().Atr(14, 100)
	t.Plans().Daily().Flush()
}

func (t *SpotTask) Clean() {
	t.Klines().Clean()
}

func (t *SpotTask) Sync() {
	t.Orders().Sync()
}
