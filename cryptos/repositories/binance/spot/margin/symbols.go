package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type SymbolsRepository struct {
	Db                 *gorm.DB
	Rdb                *redis.Client
	Ctx                context.Context
	IsolatedRepository *IsolatedRepository
}

func (r *SymbolsRepository) Isolated() *IsolatedRepository {
	if r.IsolatedRepository == nil {
		r.IsolatedRepository = &IsolatedRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.IsolatedRepository
}

func (r *SymbolsRepository) Scan() []string {
	var symbols []string
	for _, symbol := range r.Isolated().Symbols().Scan() {
		if !r.contains(symbols, symbol) {
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

func (r *SymbolsRepository) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
