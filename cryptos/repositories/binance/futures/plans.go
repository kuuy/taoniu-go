package futures

import (
  "errors"
  "fmt"
  "time"

  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type PlansRepository struct {
  Db                *gorm.DB
  SymbolsRepository *SymbolsRepository
}

type PlansInfo struct {
  Symbol    string
  Side      uint32
  Price     float32
  Quantity  float32
  Amount    float32
  Timestamp time.Time
}

func (r *PlansRepository) Ids(status int) []string {
  var ids []string
  r.Db.Model(&models.Plan{}).Select("id").Where("status", status).Find(&ids)
  return ids
}

func (r *PlansRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&models.Plan{}).Where("status", []int{0, 1}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *PlansRepository) Find(id string) (*models.Plan, error) {
  var entity *models.Plan
  result := r.Db.Take(&entity, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *PlansRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Plan{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["side"]; ok {
    query.Where("side", conditions["side"].(string))
  }
  query.Count(&total)
  return total
}

func (r *PlansRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Plan {
  var plans []*models.Plan
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "price",
    "quantity",
    "amount",
    "timestamp",
    "created_at",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["side"]; ok {
    query.Where("side", conditions["side"].(string))
  }
  query.Order("timestamp desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&plans)
  return plans
}

func (r *PlansRepository) Ranking(
  fields []string,
  conditions map[string]interface{},
  sortField string,
  sortType int,
  limit int,
) (plans []*models.Plan) {
  query := r.Db.Select(fields)
  if _, ok := conditions["interval"].(string); ok {
    query.Where("interval", conditions["interval"].(string))
  }
  if _, ok := conditions["expired_at"].(time.Time); ok {
    query.Where("created_at>?", conditions["expired_at"].(time.Time))
  }
  if sortField != "" {
    switch sortType {
    case 1:
      query.Order(fmt.Sprintf("%v ASC", sortField))
    case -1:
      query.Order(fmt.Sprintf("%v DESC", sortField))
    }
  }
  query.Limit(limit).Find(&plans)
  return
}

func (r *PlansRepository) Flush(interval string) error {
  buys, sells := r.Signals(interval)
  r.Build(interval, buys, 1)
  r.Build(interval, sells, 2)
  return nil
}

func (r *PlansRepository) Build(interval string, signals map[string]interface{}, side int) error {
  if _, ok := signals["kdj"]; !ok {
    return nil
  }
  timestamp := r.Timestamp(interval)
  for symbol, price := range signals["kdj"].(map[string]float64) {
    amount := 10.0
    if _, ok := signals["bbands"]; ok {
      if p, ok := signals["bbands"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 10.0
      }
    }
    if _, ok := signals["rsi"]; ok {
      if p, ok := signals["rsi"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 10.0
      }
    }
    if _, ok := signals["ichimoku_cloud"]; ok {
      if p, ok := signals["ichimoku_cloud"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 10.0
      }
    }
    if _, ok := signals["andean_oscillator"]; ok {
      if p, ok := signals["andean_oscillator"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 10.0
      }
    }
    if _, ok := signals["zlema"]; ok {
      if p, ok := signals["zlema"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 5.0
      }
    }
    if _, ok := signals["ha_zlema"]; ok {
      if p, ok := signals["ha_zlema"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 5.0
      }
    }
    if _, ok := signals["stoch_rsi"]; ok {
      if p, ok := signals["stoch_rsi"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 5.0
      }
    }
    if _, ok := signals["smc"]; ok {
      if p, ok := signals["smc"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 5.0
      }
    }
    if _, ok := signals["supertrend"]; ok {
      if p, ok := signals["supertrend"].(map[string]float64)[symbol]; ok {
        if p < price {
          price = p
        }
        amount += 5.0
      }
    }

    if amount < 30.0 {
      continue
    }

    tickSize, stepSize, err := r.Filters(symbol)
    if err != nil {
      continue
    }
    quantity, _ := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(price)).Float64()

    price, _ = decimal.NewFromFloat(price).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    quantity, _ = decimal.NewFromFloat(quantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    if price == 0 || quantity == 0 {
      continue
    }
    var entity models.Plan
    result := r.Db.Where("symbol=? AND interval=?", symbol, interval).Order("timestamp desc").Take(&entity)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if timestamp <= entity.Timestamp {
        continue
      }
      if side == entity.Side {
        continue
      }
    }
    entity = models.Plan{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Interval:  interval,
      Side:      side,
      Price:     price,
      Quantity:  quantity,
      Amount:    amount,
      Timestamp: timestamp,
    }
    r.Db.Create(&entity)
  }

  return nil
}

func (r *PlansRepository) Signals(interval string) (map[string]interface{}, map[string]interface{}) {
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
      "rsi",
      "ichimoku_cloud",
      "zlema",
      "ha_zlema",
      "stoch_rsi",
      "andean_oscillator",
      "smc",
      "supertrend",
    },
    interval,
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

func (r *PlansRepository) Timestep(interval string) int64 {
  switch interval {
  case "1m":
    return 60000
  case "15m":
    return 900000
  case "4h":
    return 14400000
  }
  return 86400000
}

func (r *PlansRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  switch interval {
  case "15m":
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  case "4h":
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  case "1d":
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}

func (r *PlansRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}

func (r *PlansRepository) Clean(symbol string) (err error) {
  var plan = &models.Plan{}
  for _, interval := range []string{"1m", "15m", "4h", "1d"} {
    var timestamp int64
    err = r.Db.Model(&plan).Select("timestamp").Where("symbol=? AND interval = ?", symbol, interval).Order("timestamp DESC").Offset(30).Take(&timestamp).Error
    if err == nil {
      r.Db.Where("symbol=? AND interval = ? AND timestamp < ?", symbol, interval, timestamp).Delete(&plan)
    }
  }
  r.Db.Where("status IN ?", []int{4, 5, 10}).Delete(&models.ScalpingPlan{})
  return
}
