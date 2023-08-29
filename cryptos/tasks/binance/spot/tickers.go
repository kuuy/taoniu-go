package spot

import (
  "context"
  "fmt"
  "math/rand"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersTask struct {
  Rdb                *redis.Client
  Ctx                context.Context
  Asynq              *asynq.Client
  Job                *jobs.Tickers
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func (t *TickersTask) Flush() error {
  symbols := t.Scan()
  rand.Seed(time.Now().UnixNano())
  rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    task, err := t.Job.Flush(symbols[i:j], false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_TICKERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *TickersTask) Fix() error {
  symbols := t.SymbolsRepository.Symbols()
  var items []string
  timestamp := time.Now().Unix() - 900
  whites, _ := t.Rdb.ZRangeByScore(
    t.Ctx,
    "binance:spot:tickers:flush",
    &redis.ZRangeBy{
      Min: fmt.Sprintf("%v", timestamp),
      Max: "+inf",
    },
  ).Result()
  for _, symbol := range symbols {
    if !t.contains(whites, symbol) {
      items = append(items, symbol)
    }
  }
  rand.Seed(time.Now().UnixNano())
  rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    task, err := t.Job.Flush(symbols[i:j], false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_TICKERS_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *TickersTask) FlushDelay() error {
  symbols := t.SymbolsRepository.Symbols()
  rand.Seed(time.Now().UnixNano())
  rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    task, err := t.Job.Flush(symbols[i:j], false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_TICKERS_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *TickersTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (t *TickersTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
