package spot

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type TradingsRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  ScalpingRepository *tradingsRepositories.ScalpingRepository
}

func (r *TradingsRepository) Scan() []string {
  return r.ScalpingRepository.Scan()
}
