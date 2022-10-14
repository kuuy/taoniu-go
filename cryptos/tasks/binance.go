package tasks

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"taoniu.local/cryptos/tasks/binance"
)

type BinanceTask struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (t *BinanceTask) Spot() *binance.SpotTask {
	return &binance.SpotTask{
		Db:  t.Db,
		Rdb: t.Rdb,
		Ctx: t.Ctx,
	}
}
