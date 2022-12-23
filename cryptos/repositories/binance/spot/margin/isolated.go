package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type IsolatedRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	SymbolsRepository *repositories.SymbolsRepository
	AccountRepository *repositories.AccountRepository
	OrdersRepository  *repositories.OrdersRepository
}

func (r *IsolatedRepository) Symbols() *repositories.SymbolsRepository {
	if r.SymbolsRepository == nil {
		r.SymbolsRepository = &repositories.SymbolsRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.SymbolsRepository
}

func (r *IsolatedRepository) Account() *repositories.AccountRepository {
	if r.AccountRepository == nil {
		r.AccountRepository = &repositories.AccountRepository{
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.AccountRepository
}

func (r *IsolatedRepository) Orders() *repositories.OrdersRepository {
	if r.OrdersRepository == nil {
		r.OrdersRepository = &repositories.OrdersRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.OrdersRepository
}
