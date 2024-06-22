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
  for _, symbol := range t.TradingsRepository.Scan() {
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
  for _, symbol := range t.TradingsRepository.Scan() {
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
