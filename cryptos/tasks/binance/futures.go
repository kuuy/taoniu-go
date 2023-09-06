package binance

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/futures"
)

type FuturesTask struct {
  Db             *gorm.DB
  Rdb            *redis.Client
  Ctx            context.Context
  Asynq          *asynq.Client
  CronTask       *tasks.CronTask
  AccountTask    *tasks.AccountTask
  SymbolsTask    *tasks.SymbolsTask
  TickersTask    *tasks.TickersTask
  KlinesTask     *tasks.KlinesTask
  DepthTask      *tasks.DepthTask
  OrdersTask     *tasks.OrdersTask
  IndicatorsTask *tasks.IndicatorsTask
  StrategiesTask *tasks.StrategiesTask
  PlansTask      *tasks.PlansTask
  ScalpingTask   *tasks.ScalpingTask
  TriggersTask   *tasks.TriggersTask
  TradingsTask   *tasks.TradingsTask
  AnalysisTask   *tasks.AnalysisTask
}

func (t *FuturesTask) Cron() *tasks.CronTask {
  if t.CronTask == nil {
    t.CronTask = &tasks.CronTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.CronTask
}

func (t *FuturesTask) Account() *tasks.AccountTask {
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

func (t *FuturesTask) Symbols() *tasks.SymbolsTask {
  if t.SymbolsTask == nil {
    t.SymbolsTask = &tasks.SymbolsTask{}
    t.SymbolsTask.Repository = &repositories.SymbolsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.SymbolsTask
}

func (t *FuturesTask) Tickers() *tasks.TickersTask {
  if t.TickersTask == nil {
    t.TickersTask = &tasks.TickersTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.TickersTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.TickersTask
}

func (t *FuturesTask) Klines() *tasks.KlinesTask {
  if t.KlinesTask == nil {
    t.KlinesTask = &tasks.KlinesTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.KlinesTask.Repository = &repositories.KlinesRepository{
      Db: t.Db,
    }
    t.KlinesTask.SymbolsRepository = &repositories.SymbolsRepository{
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

func (t *FuturesTask) Depth() *tasks.DepthTask {
  if t.DepthTask == nil {
    t.DepthTask = &tasks.DepthTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.DepthTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.DepthTask
}

func (t *FuturesTask) Orders() *tasks.OrdersTask {
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
    t.OrdersTask.SymbolsRepository = &repositories.SymbolsRepository{
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

func (t *FuturesTask) Indicators() *tasks.IndicatorsTask {
  if t.IndicatorsTask == nil {
    t.IndicatorsTask = &tasks.IndicatorsTask{
      Db:    t.Db,
      Asynq: t.Asynq,
    }
  }
  return t.IndicatorsTask
}

func (t *FuturesTask) Strategies() *tasks.StrategiesTask {
  if t.StrategiesTask == nil {
    t.StrategiesTask = &tasks.StrategiesTask{
      Db:    t.Db,
      Asynq: t.Asynq,
    }
  }
  return t.StrategiesTask
}

func (t *FuturesTask) Plans() *tasks.PlansTask {
  if t.PlansTask == nil {
    t.PlansTask = &tasks.PlansTask{
      Asynq: t.Asynq,
    }
  }
  return t.PlansTask
}

func (t *FuturesTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = &tasks.ScalpingTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.ScalpingTask
}

func (t *FuturesTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = &tasks.TriggersTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.TriggersTask
}

func (t *FuturesTask) Tradings() *tasks.TradingsTask {
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

func (t *FuturesTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = &tasks.AnalysisTask{
      Db: t.Db,
    }
  }
  return t.AnalysisTask
}

func (t *FuturesTask) Clean() {
  t.Klines().Clean()
}
