package plans

import (
  "context"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/plans"
  "time"
)

type Minutely struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.MinutelyRepository
}

func NewMinutely() *Minutely {
  h := &Minutely{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.MinutelyRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

func (h *Minutely) Flush(ctx context.Context, t *asynq.Task) error {
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:plans:1m:flush"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush()
  return nil
}
