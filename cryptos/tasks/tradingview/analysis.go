package tradingview

import (
  "context"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  config "taoniu.local/cryptos/config/queue"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/tradingview"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/tradingview"
)

type AnalysisTask struct {
  Rdb                *redis.Client
  Ctx                context.Context
  Asynq              *asynq.Client
  Job                *jobs.Analysis
  Repository         *repositories.AnalysisRepository
  ScalpingRepository *spotRepositories.ScalpingRepository
}

func (t *AnalysisTask) Flush() error {
  for _, symbol := range t.ScalpingRepository.Scan() {
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
