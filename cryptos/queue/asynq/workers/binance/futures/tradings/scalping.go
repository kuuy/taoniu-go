package tradings

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type Scalping struct {
  AnsqContext       *common.AnsqServerContext
  Repository        *tradingsRepositories.ScalpingRepository
  AccountRepository *repositories.AccountRepository
}

func NewScalping(ansqContext *common.AnsqServerContext) *Scalping {
  h := &Scalping{
    AnsqContext: ansqContext,
  }
  h.Repository = &tradingsRepositories.ScalpingRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  baseIndicatorsRepository := indicatorsRepositories.BaseRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.AccountRepository = &repositories.AccountRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.OrdersRepository = &repositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.PositionRepository = &repositories.PositionsRepository{
    Db: h.AnsqContext.Db,
  }
  h.Repository.AtrRepository = &indicatorsRepositories.AtrRepository{
    BaseRepository: baseIndicatorsRepository,
  }
  h.Repository.FundingRateRepository = &repositories.FundingRateRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.AccountRepository = &repositories.AccountRepository{
    Db:   h.AnsqContext.Db,
    Rdb:  h.AnsqContext.Rdb,
    Ctx:  h.AnsqContext.Ctx,
    Nats: h.AnsqContext.Nats,
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
    return nil
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
    return nil
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
