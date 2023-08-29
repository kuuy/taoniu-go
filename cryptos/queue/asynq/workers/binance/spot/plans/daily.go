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

type Daily struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.DailyRepository
}

func NewDaily() *Daily {
  h := &Daily{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.DailyRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

func (h *Daily) Flush(ctx context.Context, t *asynq.Task) error {
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:plans:1d:flush"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush()
  return nil
}
