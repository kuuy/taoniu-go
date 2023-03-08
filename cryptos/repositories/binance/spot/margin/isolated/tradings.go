package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
)

type TradingsRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	FishersRepository *repositories.FishersRepository
}

func (r *TradingsRepository) Fishers() *repositories.FishersRepository {
	if r.FishersRepository == nil {
		r.FishersRepository = &repositories.FishersRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.FishersRepository
}
