package strategies

import (
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/queue"
  models "taoniu.local/cryptos/models/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures/strategies"
)

type MinutelyTask struct {
  Db    *gorm.DB
  Asynq *asynq.Client
  Job   *jobs.Minutely
}

func (t *MinutelyTask) Atr() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Atr(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Zlema() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Zlema(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) HaZlema() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.HaZlema(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Kdj() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Kdj(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) BBands() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.BBands(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Flush() error {
  t.Atr()
  t.Zlema()
  t.HaZlema()
  t.Kdj()
  t.BBands()
  return nil
}
