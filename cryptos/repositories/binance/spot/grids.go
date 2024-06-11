package spot

import (
  "context"
  "errors"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"
  "math"
  "strconv"
  models "taoniu.local/cryptos/models/binance/spot"
)

type GridsRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *SymbolsRepository
}

func (r *GridsRepository) Symbols() *SymbolsRepository {
  if r.SymbolsRepository == nil {
    r.SymbolsRepository = &SymbolsRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.SymbolsRepository
}

func (r *GridsRepository) Flush(symbol string) error {
  var entity models.Grid
  result := r.Db.Where(
    "symbol=? AND status=1",
    symbol,
  ).Order("step desc").Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }
  context := r.Symbols().Context(symbol)
  profitTarget, _ := strconv.ParseFloat(context["profit_target"].(string), 64)
  takeProfitPrice, _ := strconv.ParseFloat(context["take_profit_price"].(string), 64)
  stopLossPoint, _ := strconv.ParseFloat(context["stop_loss_point"].(string), 64)
  if profitTarget > entity.StopLossPoint {
    price, err := r.Symbols().Price(symbol)
    if err != nil {
      return err
    }
    if price < entity.ProfitTarget {
      return nil
    }

    if entity.Step == 1 {
      return r.Close(symbol)
    }

    r.Db.Model(&entity).Update("status", 2)

    return nil
  }
  amount := entity.Amount * math.Pow(2, float64(entity.Step-1))
  entity = models.Grid{
    ID:                xid.New().String(),
    Symbol:            symbol,
    Step:              entity.Step + 1,
    Balance:           amount,
    Amount:            amount,
    ProfitTarget:      profitTarget,
    StopLossPoint:     stopLossPoint,
    TakeProfitPrice:   takeProfitPrice,
    TriggerPercent:    entity.TriggerPercent * 0.8,
    TakeProfitPercent: 0.05,
    Status:            1,
  }
  r.Db.Create(&entity)

  return nil
}

func (r *GridsRepository) Open(symbol string, amount float64) error {
  var entity models.Grid
  result := r.Db.Where(
    "symbol=? AND status=1",
    symbol,
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("grid already opened")
  }
  context := r.Symbols().Context(symbol)
  profitTarget, _ := strconv.ParseFloat(context["profit_target"].(string), 64)
  takeProfitPrice, _ := strconv.ParseFloat(context["take_profit_price"].(string), 64)
  stopLossPoint, _ := strconv.ParseFloat(context["stop_loss_point"].(string), 64)
  entity = models.Grid{
    ID:                xid.New().String(),
    Symbol:            symbol,
    Step:              1,
    Balance:           amount,
    Amount:            amount,
    ProfitTarget:      profitTarget,
    StopLossPoint:     stopLossPoint,
    TakeProfitPrice:   takeProfitPrice,
    TriggerPercent:    1,
    TakeProfitPercent: 0.05,
    Status:            1,
  }
  r.Db.Create(&entity)

  r.Rdb.SAdd(r.Ctx, "binance:spot:grids:symbols", symbol)

  return nil
}

func (r *GridsRepository) Close(symbol string) error {
  r.Db.Model(&models.Grid{}).Where(
    "symbol=? AND status=1",
    symbol,
  ).Update("status", 2)

  r.Rdb.SRem(r.Ctx, "binance:spot:grids:symbols", symbol)

  return nil
}

func (r *GridsRepository) Filter(symbol string, price float64) (*models.Grid, error) {
  var entities []*models.Grid
  r.Db.Where(
    "symbol=? AND status=1",
    symbol,
  ).Order(
    "step asc",
  ).Find(&entities)
  for _, entity := range entities {
    if price > entity.TakeProfitPrice {
      continue
    }
    if price < entity.StopLossPoint {
      continue
    }
    return entity, nil
  }

  return nil, errors.New("no valid grid")
}
