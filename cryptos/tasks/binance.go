package tasks

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  repositories "taoniu.local/cryptos/repositories/binance"
  tasks "taoniu.local/cryptos/tasks/binance"
)

type BinanceTask struct {
  Db          *gorm.DB
  Rdb         *redis.Client
  Ctx         context.Context
  Asynq       *asynq.Client
  SpotTask    *tasks.SpotTask
  FuturesTask *tasks.FuturesTask
  SavingsTask *tasks.SavingsTask
  ServerTask  *tasks.ServerTask
}

func (t *BinanceTask) Spot() *tasks.SpotTask {
  if t.SpotTask == nil {
    t.SpotTask = &tasks.SpotTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
  }
  return t.SpotTask
}

func (t *BinanceTask) Futures() *tasks.FuturesTask {
  if t.FuturesTask == nil {
    t.FuturesTask = &tasks.FuturesTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
  }
  return t.FuturesTask
}

func (t *BinanceTask) Savings() *tasks.SavingsTask {
  if t.SavingsTask == nil {
    t.SavingsTask = &tasks.SavingsTask{
      Db:  t.Db,
      Ctx: t.Ctx,
    }
  }
  return t.SavingsTask
}

func (t *BinanceTask) Server() *tasks.ServerTask {
  if t.ServerTask == nil {
    t.ServerTask = &tasks.ServerTask{
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.ServerTask.Repository = &repositories.ServerRepository{
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.ServerTask
}
