package dydx

import (
  "context"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type Tickers struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.TickersRepository
}

func NewTickers() *Tickers {
  h := &Tickers{
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.TickersRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

func (h *Tickers) Flush(ctx context.Context, t *asynq.Task) error {
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    "locks:dydx:tickers:flush",
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush()

  return nil
}

func (h *Tickers) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("dydx:tickers:flush", h.Flush)
  return nil
}
