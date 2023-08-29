package tradings

import (
  "encoding/json"
  "fmt"
  "time"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/dydx"
  dydxRepositories "taoniu.local/cryptos/repositories/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type Scalping struct {
  NatsContext *common.NatsContext
  Repository  *repositories.ScalpingRepository
}

func NewScalping(natsContext *common.NatsContext) *Scalping {
  h := &Scalping{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.ScalpingRepository{
    Db: h.NatsContext.Db,
  }
  h.Repository.MarketsRepository = &dydxRepositories.MarketsRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.AccountRepository = &dydxRepositories.AccountRepository{
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.OrdersRepository = &dydxRepositories.OrdersRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.PositionRepository = &dydxRepositories.PositionsRepository{
    Db: h.NatsContext.Db,
  }
  return h
}

func (h *Scalping) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_TRADINGS_SCALPING_PLACE, h.Place)
  return nil
}

func (h *Scalping) Place(m *nats.Msg) {
  var payload *ScalpingPlacePayload
  json.Unmarshal(m.Data, &payload)

  mutex := common.NewMutex(
    h.NatsContext.Rdb,
    h.NatsContext.Ctx,
    fmt.Sprintf("locks:dydx:tradings:scalping:place:%s", payload.PlanID),
  )
  if !mutex.Lock(30 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.PlanID)
}
