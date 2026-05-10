package futures

import (
  "errors"
  "math"
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

type PlanInfo struct {
  Symbol    string
  Side      uint32
  Price     float32
  Quantity  float32
  Amount    float32
  Timestamp time.Time
}

type signalInfo struct {
  Symbol    string
  Indicator string
  Price     float64
  Signal    int
  Timestamp int64
}

const (
  IndicatorSignalBuy  = 1
  IndicatorSignalSell = 2
)

var indicatorWeights = map[string]float64{
  "bbands":            30,
  "rsi":               20,
  "ichimoku_cloud":    20,
  "zlema":             15,
  "ha_zlema":          15,
  "stoch_rsi":         10,
  "andean_oscillator": 8,
  "smc":               8,
  "kdj":               5,
  "supertrend":        5,
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
  if symbol, ok := conditions["symbol"].(string); ok {
    query.Where("symbol", symbol)
  }
  if side, ok := conditions["side"].(uint32); ok {
    query.Where("side", side)
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
  if symbol, ok := conditions["symbol"].(string); ok {
    query.Where("symbol", symbol)
  }
  if side, ok := conditions["side"].(uint32); ok {
    query.Where("side", side)
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
  if interval, ok := conditions["interval"].(string); ok {
    query.Where("interval", interval)
  }
  if expiredAt, ok := conditions["expired_at"].(time.Time); ok {
    query.Where("created_at > ?", expiredAt)
  }
  if sortField != "" {
    switch sortType {
    case 1:
      query.Order(sortField + " ASC")
    case -1:
      query.Order(sortField + " DESC")
    }
  }
  query.Limit(limit).Find(&plans)
  return
}

func (r *PlansRepository) Flush(interval string) error {
  buys, err := r.GetSignals(interval, IndicatorSignalBuy)
  if err != nil {
    return err
  }
  sells, err := r.GetSignals(interval, IndicatorSignalSell)
  if err != nil {
    return err
  }

  if err := r.BuildPlans(interval, buys); err != nil {
    return err
  }
  if err := r.BuildPlans(interval, sells); err != nil {
    return err
  }
  return nil
}

func (r *PlansRepository) GetSignals(interval string, signal int) (map[string][]signalInfo, error) {
  timestamp := time.Now().UnixMilli()
  switch interval {
  case "1m":
    timestamp -= 900000
  case "15m":
    timestamp -= 2700000
  case "4h":
    timestamp -= 14400000
  case "1d":
    timestamp -= 86400000
  }

  indicators := make([]string, 0, len(indicatorWeights))
  for k := range indicatorWeights {
    indicators = append(indicators, k)
  }

  var strategies []*models.Strategy
  err := r.Db.Select([]string{
    "symbol",
    "indicator",
    "price",
    "signal",
    "timestamp",
  }).Where(
    "indicator IN ? AND interval = ? AND timestamp > ? AND signal = ?",
    indicators,
    interval,
    timestamp,
    signal,
  ).Order(
    "timestamp DESC",
  ).Find(&strategies).Error
  if err != nil {
    return nil, err
  }

  result := make(map[string][]signalInfo)
  for _, s := range strategies {
    result[s.Symbol] = append(result[s.Symbol], signalInfo{
      Symbol:    s.Symbol,
      Indicator: s.Indicator,
      Price:     s.Price,
      Signal:    s.Signal,
      Timestamp: s.Timestamp,
    })
  }

  return result, nil
}

func (r *PlansRepository) BuildPlans(interval string, signals map[string][]signalInfo) error {
  for symbol, indicators := range signals {
    var timestamp int64
    var n int
    for _, indicator := range indicators {
      if timestamp < indicator.Timestamp {
        timestamp = indicator.Timestamp
        n = 0
        indicators[n] = indicator
        n++
      } else if indicator.Timestamp == timestamp {
        indicators[n] = indicator
        n++
      }
    }
    indicators = indicators[:n]
    side := r.detectSide(indicators)

    basePrice, totalAmount := r.calculatePriceAndAmount(indicators)
    price, quantity, err := r.formatOrder(symbol, side, basePrice, totalAmount)
    if err != nil || price == 0 || quantity == 0 {
      continue
    }

    if r.shouldSkip(symbol, interval, timestamp, side, price, quantity, totalAmount) {
      continue
    }

    entity := models.Plan{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Interval:  interval,
      Side:      int(side),
      Price:     price,
      Quantity:  quantity,
      Amount:    totalAmount,
      Timestamp: timestamp,
    }
    r.Db.Create(&entity)
  }

  return nil
}

func (r *PlansRepository) calculatePriceAndAmount(indicators []signalInfo) (float64, float64) {
  var basePrice float64
  var totalAmount float64

  for _, ind := range indicators {
    weight, ok := indicatorWeights[ind.Indicator]
    if !ok {
      continue
    }

    if basePrice == 0 || ind.Price < basePrice {
      basePrice = ind.Price
    }
    totalAmount += weight
  }

  return basePrice, totalAmount
}

func (r *PlansRepository) formatOrder(symbol string, side uint32, price, amount float64) (float64, float64, error) {
  tickSize, stepSize, err := r.Filters(symbol)
  if err != nil {
    return 0, 0, err
  }

  if side == 1 {
    price, _ = decimal.NewFromFloat(price).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    price, _ = decimal.NewFromFloat(price).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }
  quantity, _ := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(price)).Float64()
  quantity, _ = decimal.NewFromFloat(quantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()

  return price, quantity, nil
}

func (r *PlansRepository) detectSide(indicators []signalInfo) uint32 {
  if len(indicators) == 0 {
    return 0
  }
  return uint32(indicators[0].Signal)
}

func (r *PlansRepository) shouldSkip(symbol, interval string, timestamp int64, side uint32, price, quantity, amount float64) bool {
  var entity models.Plan
  result := r.Db.Where("symbol = ? AND interval = ?", symbol, interval).
    Order("timestamp DESC").
    Take(&entity)

  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if timestamp <= entity.Timestamp {
      values := map[string]interface{}{}
      if entity.Price != price {
        if side == 1 {
          values["price"] = math.Min(entity.Price, price)
        } else {
          values["price"] = math.Max(entity.Price, price)
        }
      }
      if entity.Quantity != quantity {
        values["quantity"] = quantity
      }
      if entity.Amount != amount {
        values["amount"] = amount
      }
      if len(values) > 0 {
        r.Db.Model(&entity).Updates(values)
      }
      return true
    }
    if int(side) == entity.Side {
      values := map[string]interface{}{}
      if entity.Price != price {
        if side == 1 {
          values["price"] = math.Min(entity.Price, price)
        } else {
          values["price"] = math.Max(entity.Price, price)
        }
      }
      if entity.Quantity != quantity {
        values["quantity"] = quantity
      }
      if entity.Amount != amount {
        values["amount"] = amount
      }
      if len(values) > 0 {
        r.Db.Model(&entity).Updates(values)
      }
      return true
    }
  }
  return false
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
    minute := float64(now.Minute() / 15 * 15)
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  case "4h":
    hour := float64(now.Hour() / 4 * 4)
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
  for _, interval := range []string{"1m", "15m", "4h", "1d"} {
    var timestamp int64
    err = r.Db.Model(&models.Plan{}).Select("timestamp").
      Where("symbol = ? AND interval = ?", symbol, interval).
      Order("timestamp DESC").
      Offset(30).
      Take(&timestamp).Error
    if err == nil {
      r.Db.Where("symbol = ? AND interval = ? AND timestamp < ?", symbol, interval, timestamp).Delete(&models.Plan{})
    }
  }
  r.Db.Where("status IN ?", []int{4, 5, 10}).Delete(&models.ScalpingPlan{})
  return
}
