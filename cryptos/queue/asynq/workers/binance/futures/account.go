package futures

import (
  "context"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Account struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.AccountRepository
}

func NewAccount(ansqContext *common.AnsqServerContext) *Account {
  h := &Account{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.AccountRepository{
    Db:   h.AnsqContext.Db,
    Rdb:  h.AnsqContext.Rdb,
    Ctx:  h.AnsqContext.Ctx,
    Nats: h.AnsqContext.Nats,
  }
  return h
}

func (h *Account) Flush(ctx context.Context, t *asynq.Task) (err error) {
  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    config.LOCKS_ACCOUNT_FLUSH,
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Repository.Flush()
  return
}

func (h *Account) Register() (err error) {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_ACCOUNT_FLUSH, h.Flush)
  return
}
