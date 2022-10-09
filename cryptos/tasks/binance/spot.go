package binance

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
	tasks "taoniu.local/cryptos/tasks/binance/spot"
)

type SportTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *SportTask) Klines() *tasks.KlinesTask {
	return &tasks.KlinesTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Repository: &repositories.KlinesRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *SportTask) Indicators() *tasks.IndicatorsTask {
	return &tasks.IndicatorsTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SportTask) Strategies() *tasks.StrategiesTask {
	return &tasks.StrategiesTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *SportTask) Account() *tasks.AccountTask {
	return &tasks.AccountTask{
		Repository: &repositories.AccountRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *SportTask) Margin() *tasks.MarginTask {
	return &tasks.MarginTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}
