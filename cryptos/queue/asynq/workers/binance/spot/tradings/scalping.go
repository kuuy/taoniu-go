package tradings

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type Scalping struct {
  AnsqContext       *common.AnsqServerContext
  Repository        *repositories.ScalpingRepository
  AccountRepository *spotRepositories.AccountRepository
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
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.AccountRepository = &spotRepositories.AccountRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.PositionRepository = &spotRepositories.PositionsRepository{
    Db: h.AnsqContext.Db,
  }
  return h
}

func (h *Scalping) Place(ctx context.Context, t *asynq.Task) (err error) {
  var payload ScalpingPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_PLACE, payload.PlanId),
  )
  if !mutex.Lock(30 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.PlanId)

  return
}

func (h *Scalping) Flush(ctx context.Context, t *asynq.Task) (err error) {
  var payload ScalpingFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_FLUSH, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return
}

func (h *Scalping) Register() (err error) {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_SCALPING_PLACE, h.Place)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH, h.Flush)
  return
}
