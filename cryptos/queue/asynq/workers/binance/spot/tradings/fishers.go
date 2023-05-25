package tradings

import (
  "context"
  "encoding/json"
  fishersRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  tvRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type Fishers struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *fishersRepositories.FishersRepository
  SymbolsRepository *spotRepositories.SymbolsRepository
}

func NewFishers() *Fishers {
  h := &Fishers{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &fishersRepositories.FishersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.AnalysisRepository = &tvRepositories.AnalysisRepository{
    Db: h.Db,
  }
  h.Repository.AccountRepository = &spotRepositories.AccountRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }

  return h
}

type FishersPlacePayload struct {
  Symbol string
}

type FishersFlushPayload struct {
  Symbol string
}

func (h *Fishers) Place(ctx context.Context, t *asynq.Task) error {
  var payload FishersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Place(payload.Symbol)

  return nil
}

func (h *Fishers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload FishersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Flush(payload.Symbol)

  return nil
}
