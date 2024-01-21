package futures

import (
  "fmt"
  "slices"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type KlinesTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Klines
  Repository         *repositories.KlinesRepository
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewKlinesTask(ansqContext *common.AnsqClientContext) *KlinesTask {
  return &KlinesTask{
    AnsqContext: ansqContext,
    Repository: &repositories.KlinesRepository{
      Db: ansqContext.Db,
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

func (t *KlinesTask) Flush(interval string, limit int) error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, interval, limit, false)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_KLINES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *KlinesTask) Fix(interval string, limit int, duration int64) error {
  symbols := t.SymbolsRepository.Symbols()
  timestamp := time.Now().Unix() - duration
  whites, _ := t.AnsqContext.Rdb.ZRangeByScore(
    t.AnsqContext.Ctx,
    fmt.Sprintf(
      "binance:futures:klines:flush:%v",
      interval,
    ),
    &redis.ZRangeBy{
      Min: fmt.Sprintf("%v", timestamp),
      Max: "+inf",
    },
  ).Result()
  for _, symbol := range symbols {
    if slices.Contains(whites, symbol) {
      continue
    }
    task, err := t.Job.Flush(symbol, interval, limit, true)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_KLINES_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *KlinesTask) FlushDelay(interval string, limit int) error {
  symbols := t.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, interval, limit, true)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.BINANCE_FUTURES_KLINES_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *KlinesTask) Clean() error {
  symbols := t.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    t.Repository.Clean(symbol)
  }
  return nil
}

func (t *KlinesTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !slices.Contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}
