package plans

import (
  "context"
  "errors"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingviewRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type DailyRepository struct {
  Db                    *gorm.DB
  Rdb                   *redis.Client
  Ctx                   context.Context
  SymbolsRepository     *repositories.SymbolsRepository
  TradingviewRepository *tradingviewRepositories.AnalysisRepository
}

func (r *DailyRepository) Tradingview() *tradingviewRepositories.AnalysisRepository {
  if r.TradingviewRepository == nil {
    r.TradingviewRepository = &tradingviewRepositories.AnalysisRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.TradingviewRepository
}

func (r *DailyRepository) Symbols() *repositories.SymbolsRepository {
  if r.SymbolsRepository == nil {
    r.SymbolsRepository = &repositories.SymbolsRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.SymbolsRepository
}

func (r *DailyRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Plan{})
  if _, ok := conditions["symbols"]; ok {
    query.Where("symbol IN ?", conditions["symbols"].([]string))
  }
  query.Count(&total)
  return total
}

func (r *DailyRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Plan {
  var plans []*models.Plan
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "price",
    "quantity",
    "amount",
    "created_at",
  })
  if _, ok := conditions["symbols"]; ok {
    query.Where("symbol IN ?", conditions["symbols"].([]string))
    query.Order("timestamp desc")
  } else {
    query.Order("created_at desc")
  }
  offset := (current - 1) * pageSize
  query.Offset(offset).Limit(pageSize).Find(&plans)
  return plans
}

func (r *DailyRepository) Flush() error {
  buys, sells := r.Signals()
  r.Create(buys, 1)
  r.Create(sells, 2)
  return nil
}

func (r *DailyRepository) Fix() error {
  var plans []*models.Plan
  r.Db.Select([]string{
    "id",
    "symbol",
    "side",
  }).Order("symbol asc,updated_at desc").Find(&plans)
  var symbol string
  var side int
  for _, plan := range plans {
    if symbol == "" || symbol != plan.Symbol {
      symbol = plan.Symbol
      side = plan.Side
    } else {
      if side != plan.Side {
        side = plan.Side
      } else {
        r.Db.Delete(&plan)
      }
    }
  }
  return nil
}

func (r *DailyRepository) Create(signals map[string]interface{}, side int) error {
  if _, ok := signals["kdj"]; !ok {
    return nil
  }
  now := time.Now()
  duration := time.Hour*time.Duration(8-now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second())
  timestamp := now.Add(duration).Unix()
  for symbol, price := range signals["kdj"].(map[string]float64) {
    amount := 10.0
    if _, ok := signals["bbands"]; ok {
      if p, ok := signals["bbands"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 10
      }
    }
    if _, ok := signals["ha_zlema"]; ok {
      if p, ok := signals["ha_zlema"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 5
      }
    }
    price, quantity, _ := r.Symbols().Adjust(symbol, price, amount)
    if price == 0 || quantity == 0 {
      continue
    }
    var entity models.Plan
    result := r.Db.Where("symbol", symbol).Order("timestamp desc").Take(&entity)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if timestamp == entity.Timestamp {
        continue
      }
      if side == entity.Side {
        continue
      }
    }
    entity = models.Plan{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Side:      side,
      Price:     price,
      Quantity:  quantity,
      Amount:    amount,
      Timestamp: timestamp,
      Context:   r.Symbols().Context(symbol),
    }
    r.Db.Create(&entity)
  }

  return nil
}

func (r *DailyRepository) Signals() (map[string]interface{}, map[string]interface{}) {
  timestamp := time.Now().Unix() - 86400
  var strategies []*models.Strategy
  r.Db.Select([]string{
    "symbol",
    "indicator",
    "price",
    "signal",
  }).Where(
    "indicator in ? AND interval = ? AND timestamp > ?",
    []string{
      "kdj",
      "bbands",
      "ha_zlema",
    },
    "1d",
    timestamp,
  ).Order(
    "timestamp desc",
  ).Find(&strategies)
  var buys = make(map[string]interface{})
  var sells = make(map[string]interface{})
  for _, strategy := range strategies {
    if _, ok := buys[strategy.Indicator]; strategy.Signal == 1 && !ok {
      buys[strategy.Indicator] = make(map[string]float64)
    }
    if strategy.Signal == 1 {
      buys[strategy.Indicator].(map[string]float64)[strategy.Symbol] = strategy.Price
    }
    if _, ok := sells[strategy.Indicator]; strategy.Signal == 2 && !ok {
      sells[strategy.Indicator] = make(map[string]float64)
    }
    if strategy.Signal == 2 {
      sells[strategy.Indicator].(map[string]float64)[strategy.Symbol] = strategy.Price
    }
  }

  return buys, sells
}

func (r *DailyRepository) Filter() (*models.Plan, error) {
  now := time.Now()
  duration := time.Hour*time.Duration(8-now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second())
  timestamp := now.Add(duration).Unix()

  var entities []*models.Plan
  r.Db.Where(
    "timestamp=? AND status=0",
    timestamp,
  ).Find(&entities)
  for _, entity := range entities {
    if entity.Side != 1 {
      continue
    }
    if entity.Side == 1 && entity.Amount < 15 {
      continue
    }
    return entity, nil
  }

  return nil, errors.New("no valid plan")
}
