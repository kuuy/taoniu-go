package spot

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
  tasks "taoniu.local/cryptos/tasks/binance/spot/margin"
)

type MarginTask struct {
  Db    *gorm.DB
  Rdb   *redis.Client
  Ctx   context.Context
  Asynq *asynq.Client
}

func (t *MarginTask) Cross() *tasks.CrossTask {
  return &tasks.CrossTask{
    Db:    t.Db,
    Rdb:   t.Rdb,
    Ctx:   t.Ctx,
    Asynq: t.Asynq,
  }
}

func (t *MarginTask) Isolated() *tasks.IsolatedTask {
  return &tasks.IsolatedTask{
    Db:    t.Db,
    Rdb:   t.Rdb,
    Ctx:   t.Ctx,
    Asynq: t.Asynq,
  }
}

func (t *MarginTask) Orders() *tasks.OrdersTask {
  return &tasks.OrdersTask{
    Rdb: t.Rdb,
    Ctx: t.Ctx,
    Repository: &repositories.OrdersRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    },
  }
}

func (t *MarginTask) Flush() {
  t.Cross().Account().Flush()
  t.Isolated().Account().Flush()
  t.Isolated().Account().Liquidate()
  t.Isolated().Orders().Open()
  t.Isolated().Symbols().Flush()
  t.Orders().Flush()
}

func (t *MarginTask) Sync() {
  t.Isolated().Orders().Sync()
}
