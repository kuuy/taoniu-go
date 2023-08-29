package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  "time"
)

type Klines struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.KlinesRepository
}

func NewKlines() *Klines {
  h := &Klines{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.KlinesRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

type KlinesFlushPayload struct {
  Symbol   string
  Interval string
  Limit    int
  UseProxy bool
}

func (h *Klines) Flush(ctx context.Context, t *asynq.Task) error {
  var payload KlinesFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  if payload.UseProxy {
    h.Repository.UseProxy = true
  }

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:futures:klines:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.Interval, 0, payload.Limit)

  return nil
}

func (h *Klines) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:klines:flush", h.Flush)
  return nil
}
