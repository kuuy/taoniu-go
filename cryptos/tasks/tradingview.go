package tasks

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  jobs "taoniu.local/cryptos/queue/asynq/jobs/tradingview"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  crossTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  isolatedTradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
  repositories "taoniu.local/cryptos/repositories/tradingview"
  tasks "taoniu.local/cryptos/tasks/tradingview"
)

type TradingviewTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  Asynq        *asynq.Client
  AnalysisTask *tasks.AnalysisTask
}

func (t *TradingviewTask) Analysis() *tasks.AnalysisTask {
  if t.AnalysisTask == nil {
    t.AnalysisTask = &tasks.AnalysisTask{
      Rdb:   t.Rdb,
      Ctx:   t.Ctx,
      Asynq: t.Asynq,
    }
    t.AnalysisTask.Job = &jobs.Analysis{}
    t.AnalysisTask.Repository = &repositories.AnalysisRepository{
      Db: t.Db,
    }
    t.AnalysisTask.SymbolsRepository = &spotRepositories.SymbolsRepository{
      Db: t.Db,
    }
    t.AnalysisTask.TradingsRepository = &spotRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.AnalysisTask.TradingsRepository.FishersRepository = &tradingsRepositories.FishersRepository{
      Db: t.Db,
    }
    t.AnalysisTask.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
      Db: t.Db,
    }
    t.AnalysisTask.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.AnalysisTask.CrossTradingsRepository = &crossRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.AnalysisTask.CrossTradingsRepository.TriggersRepository = &crossTradingsRepositories.TriggersRepository{
      Db: t.Db,
    }
    t.AnalysisTask.IsolatedTradingsRepository = &isolatedRepositories.TradingsRepository{
      Db: t.Db,
    }
    t.AnalysisTask.IsolatedTradingsRepository.FishersRepository = &isolatedTradingsRepositories.FishersRepository{
      Db: t.Db,
    }
  }
  return t.AnalysisTask
}
