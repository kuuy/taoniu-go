package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type IndicatorsTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Indicators
  TradingsRepository *repositories.TradingsRepository
}

func NewIndicatorsTask(ansqContext *common.AnsqClientContext) *IndicatorsTask {
  return &IndicatorsTask{
    AnsqContext: ansqContext,
    TradingsRepository: &repositories.TradingsRepository{
      Db: ansqContext.Db,
      ScalpingRepository: &tradingsRepositories.ScalpingRepository{
        Db: ansqContext.Db,
      },
      TriggersRepository: &tradingsRepositories.TriggersRepository{
        Db: ansqContext.Db,
      },
    },
  }
}

func (t *IndicatorsTask) Pivot(interval string) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Pivot(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Atr(interval string, period int, limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Atr(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Zlema(interval string, period int, limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Zlema(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) HaZlema(interval string, period int, limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.HaZlema(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) Kdj(interval string, longPeriod int, shortPeriod int, limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Kdj(symbol, interval, longPeriod, shortPeriod, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) BBands(interval string, period int, limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.BBands(symbol, interval, period, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) IchimokuCloud(interval string) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.IchimokuCloud(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
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
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }

  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.VolumeProfile(symbol, interval, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *IndicatorsTask) AndeanOscillator(interval string, period int, length int) error {
  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }

  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.AndeanOscillator(symbol, interval, period, length, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_INDICATORS),
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
  t.IchimokuCloud(interval)
  t.VolumeProfile(interval)
  t.AndeanOscillator(interval, 90, 5)
  return nil
}
