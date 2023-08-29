package dydx

import (
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type OrdersTask struct {
  Db                 *gorm.DB
  Asynq              *asynq.Client
  Job                *jobs.Orders
  Repository         *repositories.OrdersRepository
  MarketsRepository  *repositories.MarketsRepository
  TradingsRepository *repositories.TradingsRepository
}

func (t *OrdersTask) Open() error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Open(symbol)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.DYDX_ORDERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *OrdersTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (t *OrdersTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
