package strategies

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
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type MinutelyRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository *repositories.SymbolsRepository
}

func (r *MinutelyRepository) Symbols() *repositories.SymbolsRepository {
  if r.SymbolsRepository == nil {
    r.SymbolsRepository = &repositories.SymbolsRepository{
      Db:  r.Db,
      Rdb: r.Rdb,
      Ctx: r.Ctx,
    }
  }
  return r.SymbolsRepository
}

func (r *MinutelyRepository) Atr(symbol string) error {
  day := time.Now().Format("0102")
  atrVal, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1m:%s:%s",
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

  profitTarget, _, _ := r.Symbols().Adjust(symbol, 2*price-1.5*atr, 0)
  stopLossPoint, _, _ := r.Symbols().Adjust(symbol, price-atr, 0)
  riskRewardRatio, _ := decimal.NewFromFloat(price - stopLossPoint).Div(decimal.NewFromFloat(profitTarget - price)).Round(2).Float64()
  takeProfitPrice, _, _ := r.Symbols().Adjust(symbol, stopLossPoint+(profitTarget-stopLossPoint)/2, 0)
  takeProfitRatio, _, _ := r.Symbols().Adjust(symbol, price/takeProfitPrice, 0)

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1m:%s:%s",
      symbol,
      day,
    ),
    map[string]interface{}{
      "profit_target":     profitTarget,
      "stop_loss_point":   stopLossPoint,
      "risk_reward_ratio": riskRewardRatio,
      "take_profit_price": takeProfitPrice,
      "take_profit_ratio": takeProfitRatio,
    },
  )

  return nil
}

func (r *MinutelyRepository) Zlema(symbol string) error {
  indicator := "zlema"
  interval := "1m"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1m:%s:%s",
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
  var signal int64
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
    Remark:    "",
  }
  r.Db.Create(&entity)

  return nil
}

func (r *MinutelyRepository) HaZlema(symbol string) error {
  indicator := "ha_zlema"
  interval := "1m"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1m:%s:%s",
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
  var signal int64
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
    Remark:    "",
  }
  r.Db.Create(&entity)

  return nil
}

func (r *MinutelyRepository) Kdj(symbol string) error {
  indicator := "kdj"
  interval := "1m"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1m:%s:%s",
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
  var signal int64
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
    Remark:    "",
  }
  r.Db.Create(&entity)

  return nil
}

func (r *MinutelyRepository) BBands(symbol string) error {
  indicator := "bbands"
  interval := "1m"
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1m:%s:%s",
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
  var signal int64
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
    Remark:    "",
  }
  r.Db.Create(&entity)

  return nil
}
