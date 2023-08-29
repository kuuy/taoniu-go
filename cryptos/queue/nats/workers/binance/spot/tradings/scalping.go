package tradings

import (
  "encoding/json"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
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
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  return h
}

type ScalpingPlacePayload struct {
  PlanID string `json:"plan_id"`
}

func (h *Scalping) Subscribe() error {
  return nil
}

func (h *Scalping) Place(m *nats.Msg) {
  var payload *ScalpingPlacePayload
  json.Unmarshal(m.Data, &payload)

  h.Repository.Place(payload.PlanID)
}
