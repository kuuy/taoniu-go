package margin

import (
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
)

type CrossRepository struct {
  Db                 *gorm.DB
  TradingsRepository *repositories.TradingsRepository
}

func (r *CrossRepository) Tradings() *repositories.TradingsRepository {
  if r.TradingsRepository == nil {
    r.TradingsRepository = &repositories.TradingsRepository{}
    r.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: r.Db,
    }
  }
  return r.TradingsRepository
}
