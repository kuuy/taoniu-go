package cross

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/margin/cross"
)

type Positions struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.PositionsRepository
}

func NewPositions(ansqContext *common.AnsqServerContext) *Positions {
  h := &Positions{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.PositionsRepository{
    Db: ansqContext.Db,
  }
  h.Repository.OrdersRepository = &repositories.OrdersRepository{
    Db:  ansqContext.Db,
    Ctx: ansqContext.Ctx,
  }
  return h
}

type PositionsFlushPayload struct {
  Symbol string `json:"symbol"`
  Side   int    `json:"side"`
}

func (h *Positions) Flush(ctx context.Context, t *asynq.Task) error {
  var payload PositionsFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:margin:cross:positions:flush:%s", payload.Symbol),
  )
  if !mutex.Lock(3 * time.Minute) {
    return nil
  }
  defer mutex.Unlock()

  if position, err := h.Repository.Get(payload.Symbol, payload.Side); err == nil {
    h.Repository.Flush(position)
  } else {
    h.Repository.Apply(payload.Symbol, payload.Side)
  }

  return nil
}

func (h *Positions) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:margin:cross:positions:flush", h.Flush)
  return nil
}
