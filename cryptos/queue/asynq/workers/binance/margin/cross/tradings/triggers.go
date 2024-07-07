package tradings

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/margin/cross"
  marginRepositories "taoniu.local/cryptos/repositories/binance/margin"
  crossRepositories "taoniu.local/cryptos/repositories/binance/margin/cross"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross/tradings"
)

type Triggers struct {
  AnsqContext       *common.AnsqServerContext
  Repository        *repositories.TriggersRepository
  SymbolsRepository *marginRepositories.SymbolsRepository
}

func NewTriggers(ansqContext *common.AnsqServerContext) *Triggers {
  h := &Triggers{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.TriggersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.SymbolsRepository = &marginRepositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.AccountRepository = &crossRepositories.AccountRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.OrdersRepository = &crossRepositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.PositionRepository = &crossRepositories.PositionsRepository{
    Db: h.AnsqContext.Db,
  }

  return h
}

func (h *Triggers) Place(ctx context.Context, t *asynq.Task) error {
  var payload TriggersPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_TRIGGERS_PLACE, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.ID)

  return nil
}

func (h *Triggers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload TriggersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_TRIGGERS_FLUSH, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}

func (h *Triggers) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE, h.Place)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH, h.Flush)
  return nil
}
