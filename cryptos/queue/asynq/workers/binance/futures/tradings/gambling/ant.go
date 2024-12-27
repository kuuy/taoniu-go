package gambling

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings/gambling"
)

type Ant struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.AntRepository
}

func NewAnt(ansqContext *common.AnsqServerContext) *Ant {
  h := &Ant{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.AntRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.AccountRepository = &futuresRepositories.AccountRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.OrdersRepository = &futuresRepositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.PositionRepository = &futuresRepositories.PositionsRepository{
    Db: h.AnsqContext.Db,
  }
  h.Repository.GamblingRepository = &futuresRepositories.GamblingRepository{}
  return h
}

func (h *Ant) Place(ctx context.Context, t *asynq.Task) (err error) {
  var payload AntPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_PLACE, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.ID)

  return
}

func (h *Ant) Flush(ctx context.Context, t *asynq.Task) (err error) {
  var payload AntFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_GAMBLING_ANT_FLUSH, payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return
}

func (h *Ant) Register() (err error) {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_GAMBLING_ANT_PLACE, h.Place)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_TRADINGS_GAMBLING_ANT_FLUSH, h.Flush)
  return
}
