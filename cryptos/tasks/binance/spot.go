package binance

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
	tasks "taoniu.local/cryptos/tasks/binance/spot"
)

type SpotTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *SpotTask) Tickers() *tasks.TickersTask {
	return &tasks.TickersTask{
		Db: t.Db,
		Repository: &repositories.TickersRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *SpotTask) Klines() *tasks.KlinesTask {
	return &tasks.KlinesTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SpotTask) Indicators() *tasks.IndicatorsTask {
	return &tasks.IndicatorsTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SpotTask) Strategies() *tasks.StrategiesTask {
	return &tasks.StrategiesTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SpotTask) Account() *tasks.AccountTask {
	return &tasks.AccountTask{
		Repository: &repositories.AccountRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *SpotTask) Margin() *tasks.MarginTask {
	return &tasks.MarginTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SpotTask) Analysis() *tasks.AnalysisTask {
	return &tasks.AnalysisTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SpotTask) Plans() *tasks.PlansTask {
	return &tasks.PlansTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SpotTask) Flush() {
	t.Account().Flush()
	t.Margin().Flush()
	t.Plans().Daily().Flush()
}

func (t *SpotTask) Clean() {
	t.Klines().Daily().Clean()
}
