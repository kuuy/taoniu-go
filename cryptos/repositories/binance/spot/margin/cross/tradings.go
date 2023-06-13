package cross

import (
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
)

type TradingsRepository struct {
  Db                 *gorm.DB
  TriggersRepository *repositories.TriggersRepository
}

func (r *TradingsRepository) Scan() []string {
  var symbols []string
  for _, symbol := range r.TriggersRepository.Scan() {
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
