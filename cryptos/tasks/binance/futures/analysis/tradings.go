package analysis

import (
  "gorm.io/gorm"
  repositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
  tasks "taoniu.local/cryptos/tasks/binance/futures/analysis/tradings"
)

type TradingsTask struct {
  Db           *gorm.DB
  ScalpingTask *tasks.ScalpingTask
  TriggersTask *tasks.TriggersTask
}

func (t *TradingsTask) Scalping() *tasks.ScalpingTask {
  if t.ScalpingTask == nil {
    t.ScalpingTask = &tasks.ScalpingTask{}
    t.ScalpingTask.Repository = &repositories.ScalpingRepository{
      Db: t.Db,
    }
  }
  return t.ScalpingTask
}

func (t *TradingsTask) Triggers() *tasks.TriggersTask {
  if t.TriggersTask == nil {
    t.TriggersTask = &tasks.TriggersTask{}
    t.TriggersTask.Repository = &repositories.TriggersRepository{
      Db: t.Db,
    }
  }
  return t.TriggersTask
}

func (t *TradingsTask) Flush() error {
  t.Scalping().Flush()
  t.Triggers().Flush()
  return nil
}
