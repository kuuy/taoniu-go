package gambling

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings/gambling"
)

type Scalping struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.ScalpingRepository
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
    fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_PLACE, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.ID)

  return
}

func (h *Scalping) Register() (err error) {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_GAMBLING_SCALPING_PLACE, h.Place)
  return
}
