package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	fishersRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
)

type TradingsRepository struct {
	Db                 *gorm.DB
	Rdb                *redis.Client
	Ctx                context.Context
	ScalpingRepository ScalpingRepository
	FishersRepository  *fishersRepositories.FishersRepository
}

type ScalpingRepository interface {
	Scan() []string
}

func (r *TradingsRepository) Fishers() *fishersRepositories.FishersRepository {
	if r.FishersRepository == nil {
		r.FishersRepository = &fishersRepositories.FishersRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.FishersRepository
}

func (r *TradingsRepository) Scalping() ScalpingRepository {
	return r.ScalpingRepository
}
