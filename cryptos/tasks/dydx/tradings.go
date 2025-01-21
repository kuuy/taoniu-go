package dydx

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  dydxRepositories "taoniu.local/cryptos/repositories/dydx"
  tradingsRepositories "taoniu.local/cryptos/repositories/dydx/tradings"
  tasks "taoniu.local/cryptos/tasks/dydx/tradings"
)

type TradingsTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  Asynq        *asynq.Client
  ScalpingTask *tasks.ScalpingTask
}

func (t *TradingsTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = &tasks.ScalpingTask{
      Asynq: t.Asynq,
    }
    t.ScalpingTask.Repository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.ScalpingTask.ParentRepository = &dydxRepositories.ScalpingRepository{
      Db: t.Db,
    }
  }
  return t.ScalpingTask
}
