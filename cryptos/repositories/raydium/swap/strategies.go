package swap

import (
  "context"
  "errors"
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/raydium/swap"
  models "taoniu.local/cryptos/models/raydium/swap"
)

type StrategiesRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *SymbolsRepository
  KlinesRepository  *KlinesRepository
}

func (r *StrategiesRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Strategy{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["signal"]; ok {
    query.Where("signal", conditions["signal"].(int))
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
    query.Where("signal", conditions["signal"].(int))
  }
  query.Order("timestamp desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&strategies)
  return strategies
}

func (r *StrategiesRepository) Zlema(symbol string, interval string) error {
  indicator := "zlema"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      config.REDIS_KEY_INDICATORS,
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
      config.REDIS_KEY_INDICATORS,
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

func (r *StrategiesRepository) Kdj(symbol string, interval string) error {
  indicator := "kdj"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      config.REDIS_KEY_INDICATORS,
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
      config.REDIS_KEY_INDICATORS,
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
      config.REDIS_KEY_INDICATORS,
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

func (r *StrategiesRepository) Flush(symbol string, interval string) error {
  if err := r.Zlema(symbol, interval); err != nil {
    return err
  }
  if err := r.HaZlema(symbol, interval); err != nil {
    return err
  }
  if err := r.Kdj(symbol, interval); err != nil {
    return err
  }
  if err := r.BBands(symbol, interval); err != nil {
    return err
  }
  if err := r.IchimokuCloud(symbol, interval); err != nil {
    return err
  }
  return nil
}
