package futures

import (
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/queue"
  models "taoniu.local/cryptos/models/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type OrdersTask struct {
  Db                 *gorm.DB
  Asynq              *asynq.Client
  Job                *jobs.Orders
  Repository         *repositories.OrdersRepository
  SymbolsRepository  *repositories.SymbolsRepository
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
      asynq.Queue(config.BINANCE_FUTURES_ORDERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *OrdersTask) Sync() error {
  var symbols []string
  t.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    t.Repository.Sync(symbol, 20)
  }
  return nil
}

func (t *OrdersTask) Fix() error {
  t.Repository.Fix(time.Now().Add(-30*time.Minute), 20)
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
