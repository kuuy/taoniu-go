package isolated

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
)

type TradingsRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *TradingsRepository) Scan() []string {
  var symbols []string
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
