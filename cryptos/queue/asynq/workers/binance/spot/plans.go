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

type Plans struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.PlansRepository
}

func NewPlans(ansqContext *common.AnsqServerContext) *Plans {
  h := &Plans{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.PlansRepository{
    Db: h.AnsqContext.Db,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.AnsqContext.Db,
  }
  return h
}

func (h *Plans) Flush(ctx context.Context, t *asynq.Task) error {
  var payload PlansPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:plans:%s:flush", payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Interval)

  return nil
}

func (h *Plans) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:spot:plans:flush", h.Flush)
  return nil
}
