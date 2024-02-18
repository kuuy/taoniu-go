package futures

import (
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
)

type StrategiesTask struct {
  AnsqContext *common.AnsqClientContext
  Job         *jobs.Strategies
}

func NewStrategiesTask(ansqContext *common.AnsqClientContext) *StrategiesTask {
  return &StrategiesTask{
    AnsqContext: ansqContext,
  }
}

func (t *StrategiesTask) Atr(interval string) error {
  var symbols []string
  t.AnsqContext.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Atr(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *StrategiesTask) Zlema(interval string) error {
  var symbols []string
  t.AnsqContext.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Zlema(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *StrategiesTask) HaZlema(interval string) error {
  var symbols []string
  t.AnsqContext.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.HaZlema(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *StrategiesTask) Kdj(interval string) error {
  var symbols []string
  t.AnsqContext.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.Kdj(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *StrategiesTask) BBands(interval string) error {
  var symbols []string
  t.AnsqContext.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.BBands(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *StrategiesTask) IchimokuCloud(interval string) error {
  var symbols []string
  t.AnsqContext.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    task, err := t.Job.IchimokuCloud(symbol, interval)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_STRATEGIES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *StrategiesTask) Flush(interval string) error {
  t.Atr(interval)
  t.Zlema(interval)
  t.HaZlema(interval)
  t.Kdj(interval)
  t.BBands(interval)
  t.IchimokuCloud(interval)
  return nil
}
