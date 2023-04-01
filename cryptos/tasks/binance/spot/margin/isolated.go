package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
	tasks "taoniu.local/cryptos/tasks/binance/spot/margin/isolated"
)

type IsolatedTask struct {
	Db           *gorm.DB
	Rdb          *redis.Client
	Ctx          context.Context
	Asynq        *asynq.Client
	SymbolsTask  *tasks.SymbolsTask
	AccountTask  *tasks.AccountTask
	OrdersTask   *tasks.OrdersTask
	TradingsTask *tasks.TradingsTask
}

func (t *IsolatedTask) Symbols() *tasks.SymbolsTask {
	if t.SymbolsTask == nil {
		t.SymbolsTask = &tasks.SymbolsTask{
			Repository: &repositories.SymbolsRepository{
				Db:  t.Db,
				Rdb: t.Rdb,
				Ctx: t.Ctx,
			},
		}
	}
	return t.SymbolsTask
}

func (t *IsolatedTask) Account() *tasks.AccountTask {
	if t.AccountTask == nil {
		t.AccountTask = &tasks.AccountTask{
			Repository: &repositories.AccountRepository{
				Db:  t.Db,
				Rdb: t.Rdb,
				Ctx: t.Ctx,
			},
		}
		t.AccountTask.Repository.SymbolsRepository = &repositories.SymbolsRepository{
			Db: t.Db,
		}
	}
	return t.AccountTask
}

func (t *IsolatedTask) Orders() *tasks.OrdersTask {
	if t.OrdersTask == nil {
		t.OrdersTask = &tasks.OrdersTask{
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.OrdersTask.Repository = &repositories.OrdersRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
		t.OrdersTask.Repository.Parent = &marginRepositories.OrdersRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		}
	}
	return t.OrdersTask
}

func (t *IsolatedTask) Tradings() *tasks.TradingsTask {
	if t.TradingsTask == nil {
		t.TradingsTask = &tasks.TradingsTask{
			Db:    t.Db,
			Rdb:   t.Rdb,
			Ctx:   t.Ctx,
			Asynq: t.Asynq,
		}
	}
	return t.TradingsTask
}
