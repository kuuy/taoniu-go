package spot

import (
  "context"
  "fmt"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesTask struct {
  Rdb               *redis.Client
  Ctx               context.Context
  Asynq             *asynq.Client
  Job               *jobs.Klines
  Repository        *repositories.KlinesRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func (t *KlinesTask) Flush(interval string, limit int) error {
  symbols := t.SymbolsRepository.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, interval, limit, false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_KLINES),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *KlinesTask) Fix(interval string, limit int, duration int64) error {
  symbols := t.SymbolsRepository.Symbols()
  timestamp := time.Now().Unix() - duration
  whites, _ := t.Rdb.ZRangeByScore(
    t.Ctx,
    fmt.Sprintf(
      "binance:spot:klines:flush:%v",
      interval,
    ),
    &redis.ZRangeBy{
      Min: fmt.Sprintf("%v", timestamp),
      Max: "+inf",
    },
  ).Result()
  for _, symbol := range symbols {
    if t.contains(whites, symbol) {
      continue
    }
    task, err := t.Job.Flush(symbol, interval, limit, true)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_KLINES_DELAY),
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
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_KLINES_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *KlinesTask) Clean() error {
  t.Repository.Clean()
  return nil
}

func (t *KlinesTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
