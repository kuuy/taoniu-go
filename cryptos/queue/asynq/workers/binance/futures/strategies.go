package futures

import (
  "context"
  "encoding/json"
  "fmt"
  strategiesRepositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Strategies struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.StrategiesRepository
}

type StrategyPayload struct {
  Symbol   string
  Interval string
}

func NewStrategies(ansqContext *common.AnsqServerContext) *Strategies {
  h := &Strategies{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.StrategiesRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  baseRepository := strategiesRepositories.BaseRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.Atr = &strategiesRepositories.AtrRepository{BaseRepository: baseRepository}
  h.Repository.Kdj = &strategiesRepositories.KdjRepository{BaseRepository: baseRepository}
  h.Repository.StochRsi = &strategiesRepositories.StochRsiRepository{BaseRepository: baseRepository}
  h.Repository.Zlema = &strategiesRepositories.ZlemaRepository{BaseRepository: baseRepository}
  h.Repository.HaZlema = &strategiesRepositories.HaZlemaRepository{BaseRepository: baseRepository}
  h.Repository.BBands = &strategiesRepositories.BBandsRepository{BaseRepository: baseRepository}
  h.Repository.IchimokuCloud = &strategiesRepositories.IchimokuCloudRepository{BaseRepository: baseRepository}
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.AnsqContext.Db,
  }
  return h
}

func (h *Strategies) Atr(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:strategies:atr:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:strategies:zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Zlema.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:strategies:ha_zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:strategies:kdj:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) BBands(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:strategies:bbands:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) IchimokuCloud(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:strategies:ichimoku_cloud:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.IchimokuCloud.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:futures:strategies:atr", h.Atr)
  h.AnsqContext.Mux.HandleFunc("binance:futures:strategies:zlema", h.Zlema)
  h.AnsqContext.Mux.HandleFunc("binance:futures:strategies:ha_zlema", h.HaZlema)
  h.AnsqContext.Mux.HandleFunc("binance:futures:strategies:kdj", h.Kdj)
  h.AnsqContext.Mux.HandleFunc("binance:futures:strategies:bbands", h.BBands)
  h.AnsqContext.Mux.HandleFunc("binance:futures:strategies:ichimoku_cloud", h.IchimokuCloud)
  return nil
}
