package margin

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/cross"
)

type CrossTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  Asynq        *asynq.Client
  AccountTask  *tasks.AccountTask
  TradingsTask *tasks.TradingsTask
}

func (t *CrossTask) Account() *tasks.AccountTask {
  if t.AccountTask == nil {
    t.AccountTask = &tasks.AccountTask{
      Repository: &repositories.AccountRepository{
        Rdb: t.Rdb,
        Ctx: t.Ctx,
      },
    }
  }
  return t.AccountTask
}

func (t *CrossTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = &tasks.TradingsTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
  }
  return t.TradingsTask
}
