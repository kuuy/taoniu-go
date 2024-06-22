package dydx

import (
  "encoding/json"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/dydx"
  "taoniu.local/cryptos/queue/nats/workers/dydx/tradings"
  repositories "taoniu.local/cryptos/repositories/dydx"
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
  return h
}

func (h *Scalping) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_PLANS_UPDATE, h.AddPlan)
  return nil
}

func (h *Scalping) AddPlan(m *nats.Msg) {
  var payload *PlansUpdatePayload
  json.Unmarshal(m.Data, &payload)

  if payload.Amount <= 10 {
    return
  }

  h.Repository.AddPlan(payload.ID)
  message, _ := json.Marshal(&tradings.ScalpingPlacePayload{
    PlanId: payload.ID,
  })
  h.NatsContext.Conn.Publish(config.NATS_TRADINGS_SCALPING_PLACE, message)
  h.NatsContext.Conn.Flush()
}
