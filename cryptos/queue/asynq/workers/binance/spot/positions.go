package spot

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
  return h
}

type PositionsFlushPayload struct {
  Symbol string `json:"symbol"`
}

func (h *Positions) Flush(ctx context.Context, t *asynq.Task) error {
  var payload PositionsFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:positions:flush:%s", payload.Symbol),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  if position, err := h.Repository.Get(payload.Symbol); err == nil {
    h.Repository.Flush(position)
  } else {
    h.Repository.Apply(payload.Symbol)
  }

  return nil
}

func (h *Positions) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:spot:positions:flush", h.Flush)
  return nil
}
