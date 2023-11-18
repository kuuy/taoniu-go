package tasks

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx"
  tradingsRepositories "taoniu.local/cryptos/repositories/dydx/tradings"
  tasks "taoniu.local/cryptos/tasks/dydx"
)

type DydxTask struct {
  Db             *gorm.DB
  Rdb            *redis.Client
  Ctx            context.Context
  Asynq          *asynq.Client
  TickersTask    *tasks.TickersTask
  AccountTask    *tasks.AccountTask
  MarketsTask    *tasks.MarketsTask
  OrderbookTask  *tasks.OrderbookTask
  KlinesTask     *tasks.KlinesTask
  PatternsTask   *tasks.PatternsTask
  OrdersTask     *tasks.OrdersTask
  IndicatorsTask *tasks.IndicatorsTask
  StrategiesTask *tasks.StrategiesTask
  PlansTask      *tasks.PlansTask
  TradingsTask   *tasks.TradingsTask
  ScalpingTask   *tasks.ScalpingTask
  TriggersTask   *tasks.TriggersTask
  AnalysisTask   *tasks.AnalysisTask
  ServerTask     *tasks.ServerTask
  CronTask       *tasks.CronTask
}

func (t *DydxTask) Tickers() *tasks.TickersTask {
  if t.TickersTask == nil {
    t.TickersTask = &tasks.TickersTask{
      Asynq: t.Asynq,
    }
  }
  return t.TickersTask
}

func (t *DydxTask) Account() *tasks.AccountTask {
  if t.AccountTask == nil {
    t.AccountTask = &tasks.AccountTask{
      Asynq: t.Asynq,
    }
    t.AccountTask.Repository = &repositories.AccountRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.AccountTask
}

func (t *DydxTask) Markets() *tasks.MarketsTask {
  if t.MarketsTask == nil {
    t.MarketsTask = &tasks.MarketsTask{}
    t.MarketsTask.Repository = &repositories.MarketsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.MarketsTask
}

func (t *DydxTask) Orderbook() *tasks.OrderbookTask {
  if t.OrderbookTask == nil {
    t.OrderbookTask = &tasks.OrderbookTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.OrderbookTask.MarketsRepository = &repositories.MarketsRepository{
      Db: t.Db,
    }
    t.OrderbookTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.OrderbookTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.OrderbookTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.OrderbookTask
}

func (t *DydxTask) Klines() *tasks.KlinesTask {
  if t.KlinesTask == nil {
    t.KlinesTask = &tasks.KlinesTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.KlinesTask.Repository = &repositories.KlinesRepository{
      Db: t.Db,
    }
    t.KlinesTask.MarketsRepository = &repositories.MarketsRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.KlinesTask
}

func (t *DydxTask) Patterns() *tasks.PatternsTask {
  if t.PatternsTask == nil {
    t.PatternsTask = &tasks.PatternsTask{
      Db: t.Db,
    }
  }
  return t.PatternsTask
}

func (t *DydxTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = &tasks.OrdersTask{
      Asynq: t.Asynq,
    }
    t.OrdersTask.Job = &jobs.Orders{}
    t.OrdersTask.Repository = &repositories.OrdersRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.OrdersTask.MarketsRepository = &repositories.MarketsRepository{
      Db: t.Db,
    }
    t.OrdersTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.OrdersTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.OrdersTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.OrdersTask
}

func (t *DydxTask) Indicators() *tasks.IndicatorsTask {
  if t.IndicatorsTask == nil {
    t.IndicatorsTask = &tasks.IndicatorsTask{
      Db:    t.Db,
      Asynq: t.Asynq,
    }
  }
  return t.IndicatorsTask
}

func (t *DydxTask) Strategies() *tasks.StrategiesTask {
  if t.StrategiesTask == nil {
    t.StrategiesTask = &tasks.StrategiesTask{
      Db:    t.Db,
      Asynq: t.Asynq,
    }
  }
  return t.StrategiesTask
}

func (t *DydxTask) Plans() *tasks.PlansTask {
  if t.PlansTask == nil {
    t.PlansTask = &tasks.PlansTask{
      Asynq: t.Asynq,
    }
  }
  return t.PlansTask
}

func (t *DydxTask) Tradings() *tasks.TradingsTask {
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

func (t *DydxTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = &tasks.ScalpingTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.ScalpingTask
}

func (t *DydxTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = &tasks.TriggersTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.TriggersTask
}

func (t *DydxTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = &tasks.AnalysisTask{
      Db: t.Db,
    }
  }
  return t.AnalysisTask
}

func (t *DydxTask) Cron() *tasks.CronTask {
  if t.CronTask == nil {
    t.CronTask = &tasks.CronTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.CronTask
}

func (t *DydxTask) Server() *tasks.ServerTask {
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

func (t *DydxTask) Clean() {
  t.Klines().Clean()
  t.Patterns().Candlesticks().Clean()
}
