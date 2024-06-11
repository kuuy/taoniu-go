package isolated

import (
  "context"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type TradingsRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  FishersRepository *repositories.FishersRepository
}

func (r *TradingsRepository) Scan() []string {
  var symbols []string
  for _, symbol := range r.FishersRepository.Scan() {
    if !r.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (r *TradingsRepository) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
