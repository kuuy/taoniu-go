package tasks

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	tasks "taoniu.local/cryptos/tasks/binance"
)

type BinanceTask struct {
	Db          *gorm.DB
	Rdb         *redis.Client
	Ctx         context.Context
	Asynq       *asynq.Client
	SpotTask    *tasks.SpotTask
	FuturesTask *tasks.FuturesTask
	SavingsTask *tasks.SavingsTask
}

func (t *BinanceTask) Spot() *tasks.SpotTask {
	if t.SpotTask == nil {
		t.SpotTask = &tasks.SpotTask{
			Db:    t.Db,
			Rdb:   t.Rdb,
			Ctx:   t.Ctx,
			Asynq: t.Asynq,
		}
	}
	return t.SpotTask
}

func (t *BinanceTask) Futures() *tasks.FuturesTask {
	if t.FuturesTask == nil {
		t.FuturesTask = &tasks.FuturesTask{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.FuturesTask
}

func (t *BinanceTask) Savings() *tasks.SavingsTask {
	if t.SavingsTask == nil {
		t.SavingsTask = &tasks.SavingsTask{
			Db:  t.Db,
			Ctx: t.Ctx,
		}
	}
	return t.SavingsTask
}
