package spot

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/tradings"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/spot/tradings"
)

type TradingsTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  Asynq        *asynq.Client
  FishersTask  *tasks.FishersTask
  ScalpingTask *tasks.ScalpingTask
  TriggersTask *tasks.TriggersTask
  Repository   *repositories.TradingsRepository
}

func (t *TradingsTask) Fishers() *tasks.FishersTask {
  if t.FishersTask == nil {
    t.FishersTask = &tasks.FishersTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.FishersTask.Job = &tradings.Fishers{}
    t.FishersTask.Repository = &tradingsRepositories.FishersRepository{
      Db: t.Db,
    }
  }
  return t.FishersTask
}

func (t *TradingsTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = &tasks.ScalpingTask{
      Asynq: t.Asynq,
    }
    t.ScalpingTask.Repository = &tradingsRepositories.ScalpingRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.ScalpingTask.Job = &tradings.Scalping{}
    t.ScalpingTask.Repository.SymbolsRepository = &repositories.SymbolsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.ScalpingTask.Repository.AccountRepository = &repositories.AccountRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.ScalpingTask.Repository.OrdersRepository = &repositories.OrdersRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.ScalpingTask.PlansRepository = &plansRepositories.DailyRepository{
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
    t.TriggersTask.Job = &tradings.Triggers{}
    t.TriggersTask.Repository = &tradingsRepositories.TriggersRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.TriggersTask.Repository.SymbolsRepository = &repositories.SymbolsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.TriggersTask
}

func (t *TradingsTask) Earn() error {
  return t.Repository.Earn()
}
