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

type Launchpad struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.LaunchpadRepository
}

func NewLaunchpad(ansqContext *common.AnsqServerContext) *Launchpad {
  h := &Launchpad{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.LaunchpadRepository{
    Db: h.AnsqContext.Db,
  }
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.AccountRepository = &spotRepositories.AccountRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.PositionRepository = &spotRepositories.PositionsRepository{}
  h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

func (h *Launchpad) Place(ctx context.Context, t *asynq.Task) error {
  var payload LaunchpadPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_LAUNCHPAD_PLACE, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.ID)

  return nil
}

func (h *Launchpad) Flush(ctx context.Context, t *asynq.Task) error {
  var payload LaunchpadFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_LAUNCHPAD_FLUSH, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}

func (h *Launchpad) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_LAUNCHPAD_PLACE, h.Place)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_LAUNCHPAD_FLUSH, h.Flush)
  return nil
}
