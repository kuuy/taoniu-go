package spot

import (
  "slices"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type OrdersTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Orders
  Repository         *repositories.OrdersRepository
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewOrdersTask(ansqContext *common.AnsqClientContext) *OrdersTask {
  return &OrdersTask{
    AnsqContext: ansqContext,
    Repository: &repositories.OrdersRepository{
      Db:  ansqContext.Db,
      Rdb: ansqContext.Rdb,
      Ctx: ansqContext.Ctx,
    },
    SymbolsRepository: &repositories.SymbolsRepository{
      Db: ansqContext.Db,
    },
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

func (t *OrdersTask) Open() error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Open(symbol)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_ORDERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *OrdersTask) Flush() error {
  orders := t.Repository.Gets(map[string]interface{}{})
  for _, order := range orders {
    task, err := t.Job.Flush(order.Symbol, order.OrderId)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_ORDERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *OrdersTask) Sync(startTime int64, limit int) error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Sync(symbol, startTime, limit)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_ORDERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
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
    if !slices.Contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}
