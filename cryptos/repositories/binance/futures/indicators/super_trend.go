package indicators

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
)

type SuperTrendRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}
