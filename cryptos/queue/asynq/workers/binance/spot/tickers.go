package spot

import (
  "context"
  "encoding/json"
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Tickers struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.TickersRepository
}

func NewTickers(ansqContext *common.AnsqServerContext) *Tickers {
  h := &Tickers{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.TickersRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
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

func (h *Tickers) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:spot:tickers:flush", h.Flush)
  return nil
}
