package futures

import (
  "context"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  "time"
)

type Account struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AccountRepository
}

func NewAccount() *Account {
  h := &Account{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.AccountRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

func (h *Account) Flush(ctx context.Context, t *asynq.Task) error {
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:futures:account:flush"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush()
  return nil
}

func (h *Account) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:account:flush", h.Flush)
  return nil
}
