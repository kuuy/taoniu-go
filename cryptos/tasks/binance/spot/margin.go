package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin"
)

type MarginTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *MarginTask) Isolated() *tasks.IsolatedTask {
	return &tasks.IsolatedTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}

func (t *MarginTask) Orders() *tasks.OrdersTask {
	return &tasks.OrdersTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Repository: &repositories.OrdersRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *MarginTask) Flush() {
	t.Isolated().Account().Flush()
	t.Isolated().Orders().Open()
	t.Isolated().Symbols().Flush()
	t.Isolated().Grids().Flush()
	t.Orders().Flush()
}

func (t *MarginTask) Sync() {
	t.Isolated().Orders().Sync()
}
