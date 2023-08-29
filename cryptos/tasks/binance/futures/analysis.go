package futures

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  tasks "taoniu.local/cryptos/tasks/binance/futures/analysis"
)

type AnalysisTask struct {
  Db           *gorm.DB
  Rdb          *redis.Client
  Ctx          context.Context
  TradingsTask *tasks.TradingsTask
}

func (t *AnalysisTask) Tradings() *tasks.TradingsTask {
  if t.TradingsTask == nil {
    t.TradingsTask = &tasks.TradingsTask{
      Db: t.Db,
    }
  }
  return t.TradingsTask
}

func (t *AnalysisTask) Flush() {
  t.Tradings().Flush()
}
