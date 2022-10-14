package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
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

func (t *IsolatedTask) Grids() *tasks.GridsTask {
	return &tasks.GridsTask{
		Rdb: t.Rdb,
		Ctx: t.Ctx,
		Repository: &repositories.GridsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}
