package cross

import (
  "context"
  "slices"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  repositories "taoniu.local/cryptos/repositories/binance/margin/cross/tradings"
)

type TradingsRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  ScalpingRepository *repositories.ScalpingRepository
  TriggersRepository *repositories.TriggersRepository
}

func (r *TradingsRepository) Scan() []string {
  var symbols []string
  for _, symbol := range r.ScalpingRepository.Scan() {
    if !slices.Contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range r.TriggersRepository.Scan() {
    if !slices.Contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}
