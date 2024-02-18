package tradingview

import (
  "context"
  "math/rand"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/tradingview"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  repositories "taoniu.local/cryptos/repositories/tradingview"
)

type AnalysisTask struct {
  Rdb                        *redis.Client
  Ctx                        context.Context
  Asynq                      *asynq.Client
  Job                        *jobs.Analysis
  Repository                 *repositories.AnalysisRepository
  SymbolsRepository          *spotRepositories.SymbolsRepository
  TradingsRepository         *spotRepositories.TradingsRepository
  CrossTradingsRepository    *crossRepositories.TradingsRepository
  IsolatedTradingsRepository *isolatedRepositories.TradingsRepository
}

func (t *AnalysisTask) Flush() error {
  symbols := t.Scan()
  for _, symbol := range symbols {
    task, err := t.Job.Flush("BINANCE", symbol, "1m", false)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.TRADINGVIEW_ANALYSIS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }
  return nil
}

func (t *AnalysisTask) FlushDelay() error {
  symbols := t.SymbolsRepository.Symbols()
  rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
  rnd.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
  for _, symbol := range symbols {
    task, err := t.Job.Flush("BINANCE", symbol, "1m", true)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.TRADINGVIEW_ANALYSIS_DELAY),
      asynq.MaxRetry(0),
      asynq.Timeout(10*time.Minute),
    )
  }
  return nil
}

func (t *AnalysisTask) Scan() []string {
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

func (t *AnalysisTask) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
