package spot

import (
  "context"
  "encoding/json"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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

type TickersFlushPayload struct {
  Symbols  []string
  UseProxy bool
}

func (h *Tickers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload TickersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  if payload.UseProxy {
    h.Repository.UseProxy = true
  }

  h.Repository.Flush(payload.Symbols)

  return nil
}

func (h *Tickers) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:tickers:flush", h.Flush)
  return nil
}
