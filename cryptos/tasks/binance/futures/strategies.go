package futures

import (
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  tasks "taoniu.local/cryptos/tasks/binance/futures/strategies"
)

type StrategiesTask struct {
  Db           *gorm.DB
  Asynq        *asynq.Client
  DailyTask    *tasks.DailyTask
  MinutelyTask *tasks.MinutelyTask
}

func (t *StrategiesTask) Daily() *tasks.DailyTask {
  if t.DailyTask == nil {
    t.DailyTask = &tasks.DailyTask{
      Db:    t.Db,
      Asynq: t.Asynq,
    }
  }
  return t.DailyTask
}

func (t *StrategiesTask) Minutely() *tasks.MinutelyTask {
  if t.MinutelyTask == nil {
    t.MinutelyTask = &tasks.MinutelyTask{
      Db:    t.Db,
      Asynq: t.Asynq,
    }
  }
  return t.MinutelyTask
}
