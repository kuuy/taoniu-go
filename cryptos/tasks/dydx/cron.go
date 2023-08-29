package dydx

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type CronTask struct {
  Db          *gorm.DB
  Rdb         *redis.Client
  Ctx         context.Context
  MarketsTask *MarketsTask
}

func (t *CronTask) Symbols() *MarketsTask {
  if t.MarketsTask == nil {
    t.MarketsTask = &MarketsTask{}
    t.MarketsTask.Repository = &repositories.MarketsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.MarketsTask
}

func (t *CronTask) Hourly() error {
  t.Symbols().Flush()
  return nil
}
