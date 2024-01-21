package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  "time"
)

type Klines struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.KlinesRepository
}

func NewKlines(ansqContext *common.AnsqServerContext) *Klines {
  h := &Klines{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.KlinesRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
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
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:klines:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.Interval, 0, payload.Limit)

  return nil
}

func (h *Klines) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:futures:klines:flush", h.Flush)
  return nil
}
