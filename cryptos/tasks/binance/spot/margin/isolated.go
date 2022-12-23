package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated"
)

type IsolatedTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *IsolatedTask) Symbols() *tasks.SymbolsTask {
	return &tasks.SymbolsTask{
		Repository: &repositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *IsolatedTask) Account() *tasks.AccountTask {
	return &tasks.AccountTask{
		Repository: &repositories.AccountRepository{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *IsolatedTask) Orders() *tasks.OrdersTask {
	task := &tasks.OrdersTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
	task.Repository = &repositories.OrdersRepository{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
	parentRepository := &marginRepositories.OrdersRepository{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
	task.Repository.Parent = parentRepository
	return task
}

func (t *IsolatedTask) Tradings() *tasks.TradingsTask {
	return &tasks.TradingsTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}
