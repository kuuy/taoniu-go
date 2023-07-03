package futures

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/tradings"
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
}

func (t *TradingsTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = &tasks.TriggersTask{
      Asynq: t.Asynq,
    }
    t.TriggersTask.Job = &tradings.Triggers{}
    t.TriggersTask.Repository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.TriggersTask.Repository.SymbolsRepository = &repositories.SymbolsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.TriggersTask.Repository.OrdersRepository = &repositories.OrdersRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.TriggersTask
}
