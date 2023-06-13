package tradings

import (
  "context"
  "encoding/json"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"

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
  Repository        *repositories.FishersRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewFishers() *Fishers {
  h := &Fishers{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.FishersRepository{
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
  marginRepository := &spotRepositories.MarginRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.AccountRepository = marginRepository.Isolated().Account()
  h.Repository.OrdersRepository = marginRepository.Orders()

  return h
}

type FishersFlushPayload struct {
  Symbol string
}

type FishersPlacePayload struct {
  Symbol string
}

func (h *Fishers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload FishersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Flush(payload.Symbol)

  return nil
}

func (h *Fishers) Place(ctx context.Context, t *asynq.Task) error {
  var payload FishersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Place(payload.Symbol)

  return nil
}
