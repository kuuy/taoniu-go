package spot

import (
  "context"
  "encoding/json"
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Depth struct {
  AnsqContext       *common.AnsqServerContext
  Repository        *repositories.DepthRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewDepth(ansqContext *common.AnsqServerContext) *Depth {
  h := &Depth{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.DepthRepository{
    Db: h.AnsqContext.Db,
  }
  h.SymbolsRepository = &repositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

type DepthFlushPayload struct {
  Symbol   string
  Limit    int
  UseProxy bool
}

func (h *Depth) Flush(ctx context.Context, t *asynq.Task) error {
  var payload DepthFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Flush(payload.Symbol, payload.Limit)
  h.SymbolsRepository.Slippage(payload.Symbol)

  return nil
}

func (h *Depth) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:spot:depth:flush", h.Flush)
  return nil
}
