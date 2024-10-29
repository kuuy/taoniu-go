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

type DepthTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Depth
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewDepthTask(ansqContext *common.AnsqClientContext) *DepthTask {
  return &DepthTask{
    AnsqContext: ansqContext,
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

func (t *DepthTask) Flush(limit int) error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, limit, false)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_DEPTH),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DepthTask) FlushDelay(limit int) error {
  for _, symbol := range t.TradingsRepository.Scan() {
    task, err := t.Job.Flush(symbol, limit, true)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_DEPTH),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *DepthTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !slices.Contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}
