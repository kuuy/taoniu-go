package futures

import (
  "github.com/hibiken/asynq"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/plans"
  tasks "taoniu.local/cryptos/tasks/binance/futures/plans"
)

type PlansTask struct {
  Asynq        *asynq.Client
  DailyTask    *tasks.DailyTask
  MinutelyTask *tasks.MinutelyTask
}

func (t *PlansTask) Daily() *tasks.DailyTask {
  if t.DailyTask == nil {
    t.DailyTask = &tasks.DailyTask{
      Asynq: t.Asynq,
      Job:   &jobs.Daily{},
    }
  }
  return t.DailyTask
}

func (t *PlansTask) Minutely() *tasks.MinutelyTask {
  if t.MinutelyTask == nil {
    t.MinutelyTask = &tasks.MinutelyTask{
      Asynq: t.Asynq,
      Job:   &jobs.Minutely{},
    }
  }
  return t.MinutelyTask
}
