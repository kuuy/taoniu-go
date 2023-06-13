package spot

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
)

type MarginRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  OrdersRepository   *repositories.OrdersRepository
  IsolatedRepository *repositories.IsolatedRepository
}

func (r *MarginRepository) Orders() *repositories.OrdersRepository {
  if r.OrdersRepository == nil {
    r.OrdersRepository = &repositories.OrdersRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.OrdersRepository
}

func (r *MarginRepository) Isolated() *repositories.IsolatedRepository {
  if r.IsolatedRepository == nil {
    r.IsolatedRepository = &repositories.IsolatedRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.IsolatedRepository
}
