package binance

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  crossTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  isolatedTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/spot"
)

type SpotTask struct {
  Db             *gorm.DB
  Rdb            *redis.Client
  Ctx            context.Context
  Asynq          *asynq.Client
  CronTask       *tasks.CronTask
  SymbolsTask    *tasks.SymbolsTask
  TickersTask    *tasks.TickersTask
  DepthTask      *tasks.DepthTask
  KlinesTask     *tasks.KlinesTask
  IndicatorsTask *tasks.IndicatorsTask
  StrategiesTask *tasks.StrategiesTask
  PlansTask      *tasks.PlansTask
  TradingsTask   *tasks.TradingsTask
  AccountTask    *tasks.AccountTask
  OrdersTask     *tasks.OrdersTask
  GridsTask      *tasks.GridsTask
  AnalysisTask   *tasks.AnalysisTask
  MarginTask     *tasks.MarginTask
}

func (t *SpotTask) Cron() *tasks.CronTask {
  if t.CronTask == nil {
    t.CronTask = &tasks.CronTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.CronTask
}

func (t *SpotTask) Symbols() *tasks.SymbolsTask {
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

func (t *SpotTask) Tickers() *tasks.TickersTask {
  if t.TickersTask == nil {
    t.TickersTask = &tasks.TickersTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.TickersTask.Job = &spot.Tickers{}
    t.TickersTask.Repository = &repositories.TickersRepository{
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.TickersTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository.FishersRepository = &tradingsRepositories.FishersRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.TickersTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.TickersTask.CrossTradingsRepository = &crossRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.TickersTask.CrossTradingsRepository.TriggersRepository = &crossTradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.TickersTask.IsolatedTradingsRepository = &isolatedRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.TickersTask.IsolatedTradingsRepository.FishersRepository = &isolatedTradingsRepositories.FishersRepository{
      Db: t.Db,
    }
  }
  return t.TickersTask
}

func (t *SpotTask) Depth() *tasks.DepthTask {
  if t.DepthTask == nil {
    t.DepthTask = &tasks.DepthTask{
      Asynq: t.Asynq,
    }
    t.DepthTask.Job = &spot.Depth{}
    t.DepthTask.Repository = &repositories.DepthRepository{
      Db: t.Db,
    }
    t.DepthTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository.FishersRepository = &tradingsRepositories.FishersRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.DepthTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.DepthTask.CrossTradingsRepository = &crossRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.DepthTask.CrossTradingsRepository.TriggersRepository = &crossTradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.DepthTask.IsolatedTradingsRepository = &isolatedRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.DepthTask.IsolatedTradingsRepository.FishersRepository = &isolatedTradingsRepositories.FishersRepository{
      Db: t.Db,
    }
  }
  return t.DepthTask
}

func (t *SpotTask) Klines() *tasks.KlinesTask {
  if t.KlinesTask == nil {
    t.KlinesTask = &tasks.KlinesTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.KlinesTask.Job = &spot.Klines{}
    t.KlinesTask.Repository = &repositories.KlinesRepository{
      Db: t.Db,
    }
    t.KlinesTask.SymbolsRepository = &repositories.SymbolsRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository.FishersRepository = &tradingsRepositories.FishersRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.KlinesTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.KlinesTask.CrossTradingsRepository = &crossRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.KlinesTask.CrossTradingsRepository.TriggersRepository = &crossTradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.KlinesTask.IsolatedTradingsRepository = &isolatedRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.KlinesTask.IsolatedTradingsRepository.FishersRepository = &isolatedTradingsRepositories.FishersRepository{
      Db: t.Db,
    }
  }
  return t.KlinesTask
}

func (t *SpotTask) Indicators() *tasks.IndicatorsTask {
  if t.IndicatorsTask == nil {
    t.IndicatorsTask = &tasks.IndicatorsTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.IndicatorsTask
}

func (t *SpotTask) Strategies() *tasks.StrategiesTask {
  if t.StrategiesTask == nil {
    t.StrategiesTask = &tasks.StrategiesTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.StrategiesTask
}

func (t *SpotTask) Plans() *tasks.PlansTask {
  if t.PlansTask == nil {
    t.PlansTask = &tasks.PlansTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.PlansTask
}

func (t *SpotTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = &tasks.TradingsTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.TradingsTask.Repository = &repositories.TradingsRepository{
      Db: t.Db,
    }
    t.TradingsTask.Repository.AccountRepository = &repositories.AccountRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.TradingsTask.Repository.ProductsRepository = &savingsRepositories.ProductsRepository{
      Db: t.Db,
    }
    t.TradingsTask.Repository.FishersRepository = &tradingsRepositories.FishersRepository{
      Db: t.Db,
    }
    t.TradingsTask.Repository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.TradingsTask.Repository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.TradingsTask
}

func (t *SpotTask) Account() *tasks.AccountTask {
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

func (t *SpotTask) Orders() *tasks.OrdersTask {
  if t.OrdersTask == nil {
    t.OrdersTask = &tasks.OrdersTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.OrdersTask.Repository = &repositories.OrdersRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.OrdersTask
}

func (t *SpotTask) Grids() *tasks.GridsTask {
  if t.GridsTask == nil {
    t.GridsTask = &tasks.GridsTask{
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
    t.GridsTask.Repository = &repositories.GridsRepository{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.GridsTask
}

func (t *SpotTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = &tasks.AnalysisTask{
      Db:  t.Db,
      Rdb: t.Rdb,
      Ctx: t.Ctx,
    }
  }
  return t.AnalysisTask
}

func (t *SpotTask) Margin() *tasks.MarginTask {
  if t.MarginTask == nil {
    t.MarginTask = &tasks.MarginTask{
      Db:    t.Db,
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
  }
  return t.MarginTask
}

func (t *SpotTask) Flush() {
  t.Account().Flush()
  //t.Orders().Open()
  //t.Orders().Gets()
  t.Margin().Flush()
  t.Indicators().Daily().Pivot()
  t.Indicators().Daily().Atr(14, 100)
  t.Plans().Daily().Flush()
}

func (t *SpotTask) Clean() {
  t.Klines().Clean()
}

func (t *SpotTask) Sync() {
  t.Orders().Sync()
}
