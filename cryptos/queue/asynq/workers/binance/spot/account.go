package spot

import (
  "context"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

func (h *Account) Flush(ctx context.Context, t *asynq.Task) error {
  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:account:flush"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush()
  return nil
}

func (h *Account) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:spot:account:flush", h.Flush)
  return nil
}
