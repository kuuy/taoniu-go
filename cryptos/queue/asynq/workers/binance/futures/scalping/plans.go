package scalping

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Plans struct {
  AnsqContext        *common.AnsqServerContext
  PlansRepository    *repositories.PlansRepository
  ScalpingRepository *repositories.ScalpingRepository
}

type PlansPayload struct {
  Interval string
}

func NewPlans(ansqContext *common.AnsqServerContext) *Plans {
  h := &Plans{
    AnsqContext: ansqContext,
  }
  h.PlansRepository = &repositories.PlansRepository{
    Db: h.AnsqContext.Db,
  }
  h.ScalpingRepository = &repositories.ScalpingRepository{
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
    fmt.Sprintf("locks:binance:futures:scalping:plans:%s:flush", payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  conditions := map[string]interface{}{
    "interval":   payload.Interval,
    "expired_at": time.Now().Add(-180 * time.Second),
  }
  plans := h.PlansRepository.Ranking(
    []string{"id"},
    conditions,
    "created_at",
    -1,
    100,
  )
  for _, plan := range plans {
    if !h.ScalpingRepository.IsPlanExists(plan.ID) {
      h.ScalpingRepository.AddPlan(plan.ID)
    }
  }

  return nil
}

func (h *Plans) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:futures:scalping:plans:flush", h.Flush)
  return nil
}
