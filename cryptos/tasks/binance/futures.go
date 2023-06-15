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
  Db          *gorm.DB
  Rdb         *redis.Client
  Ctx         context.Context
  Asynq       *asynq.Client
  CronTask    *tasks.CronTask
  AccountTask *tasks.AccountTask
  SymbolsTask *tasks.SymbolsTask
  TickersTask *tasks.TickersTask
  KlinesTask  *tasks.KlinesTask
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
    t.AccountTask = &tasks.AccountTask{}
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
    t.TickersTask.Job = &jobs.Tickers{}
    t.TickersTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository = &repositories.TradingsRepository{
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
    t.KlinesTask.Job = &jobs.Klines{}
    t.KlinesTask.Repository = &repositories.KlinesRepository{
      Db: t.Db,
    }
    t.KlinesTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.KlinesTask
}

func (t *FuturesTask) Flush() {
  t.Account().Flush()
  //t.Indicators().Daily().Pivot()
  //t.Indicators().Daily().Atr(14, 100)
  //t.Plans().Daily().Flush()
}

func (t *FuturesTask) Clean() {
  t.Klines().Clean()
}

func (t *FuturesTask) Sync() {
}
