package dydx

import (
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "time"

  config "taoniu.local/cryptos/config/queue"
  models "taoniu.local/cryptos/models/dydx"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx"
)

type IndicatorsTask struct {
  Db    *gorm.DB
  Asynq *asynq.Client
  Job   *jobs.Indicators
}

func (t *IndicatorsTask) Pivot(interval string) error {
  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Pivot(symbol, interval)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Atr(interval string, period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Atr(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Zlema(interval string, period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Zlema(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) HaZlema(interval string, period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.HaZlema(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Kdj(interval string, longPeriod int, shortPeriod int, limit int) error {
  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Kdj(symbol, interval, longPeriod, shortPeriod, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) BBands(interval string, period int, limit int) error {
  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.BBands(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) VolumeProfile(interval string) error {
  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }

  var symbols []string
  t.Db.Model(models.Market{}).Select("symbol").Where("status=?", "ONLINE").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.VolumeProfile(symbol, interval, limit)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Flush(interval string) error {
  t.Pivot(interval)
  t.Atr(interval, 14, 100)
  t.Zlema(interval, 14, 100)
  t.HaZlema(interval, 14, 100)
  t.Kdj(interval, 9, 3, 100)
  t.BBands(interval, 14, 100)
  t.VolumeProfile(interval)
  return nil
}
