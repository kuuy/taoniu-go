package tor

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
)

type ProxiesRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  BridgesRepository *BridgesRepository
}

func (r *ProxiesRepository) Start(id int) error {
  limit := 20
  bridges, err := r.BridgesRepository.Random(id, limit)
  if err != nil {
    return err
  }
  return r.BridgesRepository.Monitor(id, bridges, false)
}
