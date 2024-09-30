package spot

import (
  "encoding/json"
  "github.com/nats-io/nats.go"
  "log"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Plans struct {
  NatsContext *common.NatsContext
  Repository  *repositories.PlansRepository
}

func NewPlans(natsContext *common.NatsContext) *Plans {
  h := &Plans{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.PlansRepository{
    Db: h.NatsContext.Db,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.NatsContext.Db,
  }
  return h
}

func (h *Plans) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_STRATEGIES_UPDATE, h.Create)
  return nil
}

func (h *Plans) Create(m *nats.Msg) {
  var payload *StrategiesUpdatePayload
  json.Unmarshal(m.Data, &payload)

  plan, err := h.Repository.Create(payload.Symbol, payload.Interval)
  if err != nil {
    log.Println("plan create error", err.Error())
    return
  }
  message, _ := json.Marshal(&PlansUpdatePayload{
    ID:     plan.ID,
    Side:   plan.Side,
    Amount: plan.Amount,
  })
  h.NatsContext.Conn.Publish(config.NATS_PLANS_UPDATE, message)
  h.NatsContext.Conn.Flush()
}
