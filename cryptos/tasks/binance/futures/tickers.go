package futures

import (
  "math/rand"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  jobs "taoniu.local/cryptos/queue/asynq/jobs/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type TickersTask struct {
  AnsqContext        *common.AnsqClientContext
  Job                *jobs.Tickers
  TradingsRepository *repositories.TradingsRepository
}

func NewTickersTask(ansqContext *common.AnsqClientContext) *TickersTask {
  return &TickersTask{
    AnsqContext: ansqContext,
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

func (t *TickersTask) Flush() error {
  symbols := t.TradingsRepository.Scan()
  rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
  rnd.Shuffle(len(symbols), func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })
  for i := 0; i < len(symbols); i += 20 {
    j := i + 20
    if j > len(symbols) {
      j = len(symbols)
    }
    task, err := t.Job.Flush(symbols[i:j], false)
    if err != nil {
      return err
    }
    t.AnsqContext.Conn.Enqueue(
      task,
      asynq.Queue(config.ASYNQ_QUEUE_TICKERS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}
