package dydx

import (
  "errors"
  "time"

  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/dydx"
)

type PlansRepository struct {
  Db                *gorm.DB
  MarketsRepository *MarketsRepository
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
  result := r.Db.First(&entity, "id=?", id)
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
    query.Where("side", conditions["int"].(string))
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
      "ha_zlema",
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

func (r *PlansRepository) Create(symbol string, interval string) (plan models.Plan, err error) {
  var strategy models.Strategy
  result := r.Db.Select([]string{"price", "signal", "timestamp"}).Where(
    "symbol = ? AND interval = ? AND indicator = 'kdj' AND timestamp >= ?",
    symbol,
    interval,
    r.Timestamp(interval)-60000,
  ).Order(
    "timestamp desc",
  ).Take(&strategy)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = result.Error
    return
  }

  price := strategy.Price
  amount := 10.0

  for _, indicators := range [][]string{
    {"bbands"},
    {"zlema", "ha_zlema"},
  } {
    var entity models.Strategy
    r.Db.Select([]string{
      "indicator",
      "price",
      "signal",
    }).Where(
      "symbol = ? AND interval = ? AND indicator IN ? AND timestamp = ?",
      symbol,
      interval,
      indicators,
      strategy.Timestamp-14*r.Timestep(interval),
    ).Order(
      "timestamp desc",
    ).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      continue
    }
    if entity.Signal != strategy.Signal {
      continue
    }
    if entity.Indicator == "bbands" {
      amount += 10
    }
    if entity.Indicator == "zlema" || entity.Indicator == "ha_zlema" {
      amount += 5
    }
  }

  tickSize, stepSize, err := r.Filters(symbol)
  if err != nil {
    return
  }
  price, _ = decimal.NewFromFloat(price).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  quantity, _ := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(price)).Float64()
  quantity, _ = decimal.NewFromFloat(quantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()

  var entity models.Plan
  result = r.Db.Where("symbol=? AND interval=? AND timestamp=?", symbol, interval, strategy.Timestamp).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("plan timestamp has been taken")
    return
  }
  plan = models.Plan{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Interval:  interval,
    Side:      strategy.Signal,
    Price:     price,
    Quantity:  quantity,
    Amount:    amount,
    Timestamp: strategy.Timestamp,
  }
  r.Db.Create(&plan)

  return
}

func (r *PlansRepository) Timestep(interval string) int64 {
  if interval == "1m" {
    return 60000
  } else if interval == "15m" {
    return 900000
  } else if interval == "4h" {
    return 14400000
  }
  return 86400000
}

func (r *PlansRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  if interval == "1m" {
    duration = duration - time.Minute
  } else if interval == "15m" {
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  } else if interval == "4h" {
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  } else {
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}

func (r *PlansRepository) Filters(symbol string) (float64, stepSize float64, err error) {
  entity, err := r.MarketsRepository.Get(symbol)
  if err != nil {
    return
  }
  return entity.TickSize, entity.StepSize, nil
}
