package spot

import (
  "math/rand"
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type DepthTask struct {
  Asynq                      *asynq.Client
  Job                        *jobs.Depth
  Repository                 *repositories.DepthRepository
  SymbolsRepository          *repositories.SymbolsRepository
  TradingsRepository         *repositories.TradingsRepository
  CrossTradingsRepository    *crossRepositories.TradingsRepository
  IsolatedTradingsRepository *isolatedRepositories.TradingsRepository
}

func (t *DepthTask) Flush() error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_DEPTH),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *DepthTask) FlushDelay() error {
  symbols := t.SymbolsRepository.Symbols()
  rand.Seed(time.Now().UnixNano())
  rand.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
  for _, symbol := range symbols {
    task, err := t.Job.Flush(symbol, true)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.BINANCE_SPOT_DEPTH_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}

func (t *DepthTask) Scan() []string {
  var symbols []string
  for _, symbol := range t.TradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range t.CrossTradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range t.IsolatedTradingsRepository.Scan() {
    if !t.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (h *DepthTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
