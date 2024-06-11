package tradings

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type Scalping struct {
  AnsqContext       *common.AnsqServerContext
  Repository        *repositories.ScalpingRepository
  AccountRepository *futuresRepositories.AccountRepository
}

func NewScalping(ansqContext *common.AnsqServerContext) *Scalping {
  h := &Scalping{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.ScalpingRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.AccountRepository = &futuresRepositories.AccountRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.OrdersRepository = &futuresRepositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.PositionRepository = &futuresRepositories.PositionsRepository{
    Db: h.AnsqContext.Db,
  }
  h.AccountRepository = &futuresRepositories.AccountRepository{
    Db:   h.AnsqContext.Db,
    Rdb:  h.AnsqContext.Rdb,
    Ctx:  h.AnsqContext.Ctx,
    Nats: h.AnsqContext.Nats,
  }
  return h
}

func (h *Scalping) Place(ctx context.Context, t *asynq.Task) error {
  var payload ScalpingPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_PLACE, payload.PlanID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }

  err := h.Repository.Place(payload.PlanID)
  if err != nil {
    mutex.Unlock()
  }

  return nil
}

func (h *Scalping) Flush(ctx context.Context, t *asynq.Task) error {
  var payload ScalpingFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_FLUSH, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}

func (h *Scalping) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_SCALPING_PLACE, h.Place)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH, h.Flush)
  return nil
}
