package cross

import (
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/margin/cross"
)

type Account struct{}

func (h *Account) Flush() (*asynq.Task, error) {
  return asynq.NewTask(config.ASYNQ_JOBS_ACCOUNT_FLUSH, nil), nil
}
