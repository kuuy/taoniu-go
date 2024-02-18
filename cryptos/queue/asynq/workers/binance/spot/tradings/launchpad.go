package tradings

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type Launchpad struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.LaunchpadRepository
}

func NewLaunchpad() *Launchpad {
  h := &Launchpad{
    Db:  common.NewDB(1),
    Rdb: common.NewRedis(1),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.LaunchpadRepository{
    Db: h.Db,
  }
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.AccountRepository = &spotRepositories.AccountRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.PositionRepository = &spotRepositories.PositionsRepository{}
  h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

type LaunchpadPlacePayload struct {
  ID string
}

type LaunchpadFlushPayload struct {
  ID string
}

func (h *Launchpad) Place(ctx context.Context, t *asynq.Task) error {
  var payload LaunchpadPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:tradings:launchpad:place:%s", payload.ID),
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
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:tradings:launchpad:flush:%s", payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}
