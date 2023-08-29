package strategies

import (
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/queue"
  models "taoniu.local/cryptos/models/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/strategies"
)

type DailyTask struct {
  Db    *gorm.DB
  Asynq *asynq.Client
  Job   *jobs.Daily
}

func (t *DailyTask) Atr() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Atr(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DailyTask) Zlema() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Zlema(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DailyTask) HaZlema() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.HaZlema(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DailyTask) Kdj() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Kdj(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DailyTask) BBands() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.BBands(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DailyTask) Flush() error {
  t.Atr()
  t.Zlema()
  t.HaZlema()
  t.Kdj()
  t.BBands()
  return nil
}
