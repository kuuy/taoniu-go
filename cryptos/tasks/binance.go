package tasks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance"
	"taoniu.local/cryptos/tasks/binance"
)

type BinanceTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *BinanceTask) Symbols() *binance.SymbolsTask {
	return &binance.SymbolsTask{
		Repository: &repositories.SymbolsRepository{
			Db:  t.Db,
			Rdb: t.Rdb,
			Ctx: t.Ctx,
		},
	}
}

func (t *BinanceTask) Spot() *binance.SpotTask {
	return &binance.SpotTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}
