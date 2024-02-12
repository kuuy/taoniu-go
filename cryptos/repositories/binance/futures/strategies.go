package futures

import (
  "context"
  "errors"
  "fmt"
  "github.com/shopspring/decimal"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type StrategiesRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *SymbolsRepository
}

func (r *StrategiesRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Plan{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["signal"]; ok {
    query.Where("signal", conditions["signal"].(string))
  }
  query.Count(&total)
  return total
}

func (r *StrategiesRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Strategy {
  var strategies []*models.Strategy
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "indicator",
    "signal",
    "price",
    "timestamp",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["signal"]; ok {
    query.Where("signal", conditions["signal"].(string))
  }
  query.Order("timestamp desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&strategies)
  return strategies
}

func (r *StrategiesRepository) Atr(symbol string, interval string) error {
  tickSize, _, err := r.Filters(symbol)
  if err != nil {
    return err
  }

  day := time.Now().Format("0102")
  atrVal, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      day,
    ),
    "atr",
  ).Result()
  if err != nil {
    return err
  }
  atr, _ := strconv.ParseFloat(atrVal, 64)
  priceVal, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:realtime:%s",
      symbol,
    ),
    "price",
  ).Result()
  if err != nil {
    return err
  }
  price, _ := strconv.ParseFloat(priceVal, 64)

  profitTarget, _ := decimal.NewFromFloat(price).Mul(decimal.NewFromInt(2)).Sub(
    decimal.NewFromFloat(atr).Mul(decimal.NewFromFloat(1.5)),
  ).Float64()
  stopLossPoint, _ := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(atr)).Float64()
  takeProfitPrice, _ := decimal.NewFromFloat(stopLossPoint).Add(
    decimal.NewFromFloat(profitTarget).Sub(decimal.NewFromFloat(stopLossPoint)).Div(decimal.NewFromInt(2)),
  ).Float64()
  riskRewardRatio, _ := decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(stopLossPoint)).Div(
    decimal.NewFromFloat(profitTarget).Sub(decimal.NewFromFloat(price)),
  ).Round(2).Float64()
  takeProfitRatio, _ := decimal.NewFromFloat(price).Div(decimal.NewFromFloat(takeProfitPrice)).Round(2).Float64()

  profitTarget, _ = decimal.NewFromFloat(profitTarget).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  stopLossPoint, _ = decimal.NewFromFloat(stopLossPoint).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  takeProfitPrice, _ = decimal.NewFromFloat(takeProfitPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      day,
    ),
    map[string]interface{}{
      "profit_target":     profitTarget,
      "stop_loss_point":   stopLossPoint,
      "take_profit_price": takeProfitPrice,
      "risk_reward_ratio": riskRewardRatio,
      "take_profit_ratio": takeProfitRatio,
    },
  )

  return nil
}

func (r *StrategiesRepository) Zlema(symbol string, interval string) error {
  indicator := "zlema"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      time.Now().Format("0102"),
    ),
    indicator,
  ).Result()
  if err != nil {
    return err
  }
  data := strings.Split(val, ",")

  price, _ := strconv.ParseFloat(data[2], 64)
  zlema1, _ := strconv.ParseFloat(data[0], 64)
  zlema2, _ := strconv.ParseFloat(data[1], 64)
  timestamp, _ := strconv.ParseInt(data[3], 10, 64)
  if zlema1*zlema2 >= 0.0 {
    return nil
  }
  var signal int
  if zlema2 > 0 {
    signal = 1
  } else {
    signal = 2
  }
  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    indicator,
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return nil
    }
    if entity.Timestamp >= timestamp {
      return nil
    }
  }
  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: indicator,
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)

  return nil
}

func (r *StrategiesRepository) HaZlema(symbol string, interval string) error {
  indicator := "ha_zlema"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      time.Now().Format("0102"),
    ),
    "ha_zlema",
  ).Result()
  if err != nil {
    return err
  }
  data := strings.Split(val, ",")

  price, _ := strconv.ParseFloat(data[2], 64)
  zlema1, _ := strconv.ParseFloat(data[0], 64)
  zlema2, _ := strconv.ParseFloat(data[1], 64)
  timestamp, _ := strconv.ParseInt(data[3], 10, 64)
  if zlema1*zlema2 >= 0.0 {
    return nil
  }
  var signal int
  if zlema2 > 0 {
    signal = 1
  } else {
    signal = 2
  }
  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    indicator,
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return nil
    }
    if entity.Timestamp >= timestamp {
      return nil
    }
  }
  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: indicator,
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)

  return nil
}

func (r *StrategiesRepository) Kdj(symbol string, interval string) error {
  indicator := "kdj"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      time.Now().Format("0102"),
    ),
    indicator,
  ).Result()
  if err != nil {
    return err
  }
  data := strings.Split(val, ",")

  k, _ := strconv.ParseFloat(data[0], 64)
  d, _ := strconv.ParseFloat(data[1], 64)
  j, _ := strconv.ParseFloat(data[2], 64)
  price, _ := strconv.ParseFloat(data[3], 64)
  timestamp, _ := strconv.ParseInt(data[4], 10, 64)
  var signal int
  if k < 20 && d < 30 && j < 60 {
    signal = 1
  }
  if k > 80 && d > 70 && j > 90 {
    signal = 2
  }
  if signal == 0 {
    return nil
  }
  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    indicator,
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return nil
    }
    if entity.Timestamp >= timestamp {
      return nil
    }
  }
  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: indicator,
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)

  return nil
}

func (r *StrategiesRepository) BBands(symbol string, interval string) error {
  indicator := "bbands"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      time.Now().Format("0102"),
    ),
    indicator,
  ).Result()
  if err != nil {
    return err
  }
  data := strings.Split(val, ",")

  b1, _ := strconv.ParseFloat(data[0], 64)
  b2, _ := strconv.ParseFloat(data[1], 64)
  b3, _ := strconv.ParseFloat(data[2], 64)
  w1, _ := strconv.ParseFloat(data[3], 64)
  w2, _ := strconv.ParseFloat(data[4], 64)
  w3, _ := strconv.ParseFloat(data[5], 64)
  price, _ := strconv.ParseFloat(data[6], 64)
  timestamp, _ := strconv.ParseInt(data[7], 10, 64)
  var signal int
  if b1 < 0.5 && b2 < 0.5 && b3 > 0.5 {
    signal = 1
  }
  if b1 > 0.5 && b2 < 0.5 && b3 < 0.5 {
    signal = 2
  }
  if b1 > 0.8 && b2 > 0.8 && b3 > 0.8 {
    signal = 1
  }
  if b1 > 0.8 && b2 > 0.8 && b3 < 0.8 {
    signal = 2
  }
  if w1 < 0.1 && w2 < 0.1 && w3 < 0.1 {
    if w1 < 0.03 || w2 < 0.03 || w3 > 0.03 {
      return nil
    }
  }
  if signal == 0 {
    return nil
  }
  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    indicator,
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return nil
    }
    if entity.Timestamp >= timestamp {
      return nil
    }
  }
  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: indicator,
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return nil
}

func (r *StrategiesRepository) IchimokuCloud(symbol string, interval string) error {
  indicator := "ichimoku_cloud"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:%s:%s:%s",
      interval,
      symbol,
      time.Now().Format("0102"),
    ),
    indicator,
  ).Result()
  if err != nil {
    return err
  }
  data := strings.Split(val, ",")

  signal, _ := strconv.Atoi(data[0])
  price, _ := strconv.ParseFloat(data[6], 64)
  timestamp, _ := strconv.ParseInt(data[7], 10, 64)

  if signal == 0 {
    return nil
  }
  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    indicator,
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return nil
    }
    if entity.Timestamp >= timestamp {
      return nil
    }
  }
  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: indicator,
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return nil
}

func (r *StrategiesRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}
