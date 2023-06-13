package cross

import (
  "context"
  "github.com/hibiken/asynq"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  tradingsJobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/margin/cross/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin/cross/tradings"
)

type TradingsTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  Asynq        *asynq.Client
  TriggersTask *tasks.TriggersTask
}

func (t *TradingsTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = &tasks.TriggersTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.TriggersTask.Job = &tradingsJobs.Triggers{}
    t.TriggersTask.Repository = &repositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.TriggersTask
}
