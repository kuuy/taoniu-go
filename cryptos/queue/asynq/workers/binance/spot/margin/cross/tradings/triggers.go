package tradings

import (
  "context"
  "encoding/json"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
)

type Triggers struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *repositories.TriggersRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewTriggers() *Triggers {
  h := &Triggers{
    Db:  common.NewDB(1),
    Rdb: common.NewRedis(1),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.TriggersRepository{
    Db: h.Db,
  }
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.AccountRepository = &crossRepositories.AccountRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.MarginAccountRepository = &marginRepositories.AccountRepository{}
  h.Repository.OrdersRepository = &marginRepositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }

  return h
}

type TriggersFlushPayload struct {
  Symbol string
}

type TriggersPlacePayload struct {
  Symbol string
}

func (h *Triggers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload TriggersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Flush(payload.Symbol)

  return nil
}

func (h *Triggers) Place(ctx context.Context, t *asynq.Task) error {
  var payload TriggersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Place(payload.Symbol)

  return nil
}
