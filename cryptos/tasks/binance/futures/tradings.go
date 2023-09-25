package futures

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/futures/tradings"
)

type TradingsTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  Asynq        *asynq.Client
  TriggersTask *tasks.TriggersTask
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
    t.ScalpingTask.ParentRepository = &repositories.ScalpingRepository{
      Db: t.Db,
    }
  }
  return t.ScalpingTask
}

func (t *TradingsTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = &tasks.TriggersTask{
      Asynq: t.Asynq,
    }
    t.TriggersTask.Repository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.TriggersTask
}
