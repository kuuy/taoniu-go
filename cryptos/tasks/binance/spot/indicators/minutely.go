package indicators

import (
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/queue"
  models "taoniu.local/cryptos/models/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot/indicators"
  repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type MinutelyTask struct {
  Db         *gorm.DB
  Asynq      *asynq.Client
  Job        *jobs.Minutely
  Repository *repositories.MinutelyRepository
}

func (t *MinutelyTask) Pivot() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Pivot(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Atr(period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Atr(symbol, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Zlema(period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Zlema(symbol, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) HaZlema(period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.HaZlema(symbol, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Kdj(longPeriod int, shortPeriod int, limit int) error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Kdj(symbol, longPeriod, shortPeriod, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) BBands(period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.BBands(symbol, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) VolumeProfile(limit int) error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.VolumeProfile(symbol, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *MinutelyTask) Flush() error {
  t.Pivot()
  t.Atr(14, 100)
  t.Zlema(14, 100)
  t.HaZlema(14, 100)
  t.Kdj(9, 3, 100)
  t.BBands(14, 100)
  t.VolumeProfile(1440)
  return nil
}
