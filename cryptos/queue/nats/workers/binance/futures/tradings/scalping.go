package tradings

import (
  "encoding/json"
  "fmt"
  "time"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type Scalping struct {
  NatsContext *common.NatsContext
  Repository  *tradingsRepositories.ScalpingRepository
}

func NewScalping(natsContext *common.NatsContext) *Scalping {
  h := &Scalping{
    NatsContext: natsContext,
  }
  h.Repository = &tradingsRepositories.ScalpingRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  baseIndicatorsRepository := indicatorsRepositories.BaseRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.AccountRepository = &repositories.AccountRepository{
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.OrdersRepository = &repositories.OrdersRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.PositionRepository = &repositories.PositionsRepository{
    Db: h.NatsContext.Db,
  }
  h.Repository.AtrRepository = &indicatorsRepositories.AtrRepository{
    BaseRepository: baseIndicatorsRepository,
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
    fmt.Sprintf("locks:binance:futures:tradings:scalping:place:%s", payload.PlanId),
  )
  if !mutex.Lock(30 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.PlanId)
}
