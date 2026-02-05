package futures

import (
  "context"
  "errors"
  "fmt"
  "math"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/binance/futures"
  "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type StopLossRepository struct {
  Db                      *gorm.DB
  Rdb                     *redis.Client
  Ctx                     context.Context
  SymbolsRepository       *SymbolsRepository
  TickersRepository       *TickersRepository
  AtrRepository           *indicators.AtrRepository
  VolumeProfileRepository *indicators.VolumeProfileRepository
  IndicatorsRepository    *IndicatorsRepository
}

func (r *StopLossRepository) Get(symbol string, side int) (result *StopLossInfo, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(config.REDIS_KEY_STOPLOSS, side, symbol, day)
  data, err := r.Rdb.HGetAll(r.Ctx, redisKey).Result()
  if err != nil {
    return
  }
  if len(data) == 0 {
    err = errors.New("no stop loss found")
    return
  }
  result = &StopLossInfo{Symbol: symbol}
  result.Side, _ = strconv.Atoi(data["side"])
  result.EntryPrice, _ = strconv.ParseFloat(data["entry_price"], 64)
  result.InitialStop, _ = strconv.ParseFloat(data["initial_stop"], 64)
  result.ActiveStop, _ = strconv.ParseFloat(data["active_stop"], 64)
  result.TakeProfit1, _ = strconv.ParseFloat(data["take_profit_1"], 64)
  result.TakeProfit2, _ = strconv.ParseFloat(data["take_profit_2"], 64)
  result.ATR, _ = strconv.ParseFloat(data["atr"], 64)
  result.StopType = data["stop_type"]
  return
}

func (r *StopLossRepository) Calc(symbol string, side int, entryPrice float64, currentStop float64, leverage int, risk float64) (result *StopLossInfo, err error) {
  result = &StopLossInfo{
    Symbol:      symbol,
    Side:        side,
    EntryPrice:  entryPrice,
    Leverage:    leverage,
    Risk:        risk,
    ShouldTrade: true,
  }

  if currentStop == 0.0 {
    result.CurrentPrice = entryPrice
  } else {
    var price float64
    price, err = r.TickersRepository.Price(symbol)
    if err != nil {
      return nil, fmt.Errorf("tickers repository price error: %v", err)
    }
    result.CurrentPrice = price
  }

  atr, err := r.AtrRepository.Get(symbol, "15m")
  if err != nil {
    if currentStop == 0 {
      atr = 0
    } else {
      return nil, fmt.Errorf("atr repository get error: %v", err)
    }
  }
  result.ATR = atr
  result.ATRMultiplier = r.AtrRepository.Multiplier(entryPrice, atr)

  poc, vah, val, err := r.VolumeProfileRepository.Get(symbol, "15m")
  if err != nil {
    poc, vah, val = 0, 0, 0
  }

  tickSize, _, err := r.Filters(symbol)
  if err != nil {
    return
  }

  if side == 1 { // LONG
    result = r.calculateLongStops(result, poc, vah, val, tickSize)
    result = r.updateLongTrailing(result, currentStop, tickSize)
  } else { // SHORT
    result = r.calculateShortStops(result, poc, vah, val, tickSize)
    result = r.updateShortTrailing(result, currentStop, tickSize)
  }

  if result.RiskReward < 1.2 {
    result.ShouldTrade = false
    result.RejectReason = fmt.Sprintf("risk reward ratio less than %.2f", result.RiskReward)
  }

  stopDistance := math.Abs(entryPrice-result.InitialStop) / entryPrice
  if stopDistance > 0.05 {
    result.ShouldTrade = false
    result.RejectReason = fmt.Sprintf("stop distance more than %.2f%%", stopDistance*100)
  }
  if stopDistance < 0.002 {
    result.ShouldTrade = false
    result.RejectReason = fmt.Sprintf("stop distance less than %.2f%%", stopDistance*100)
  }

  result.ActiveStop = result.InitialStop
  result.StopType = "initial"

  return
}

func (r *StopLossRepository) calculateLongStops(result *StopLossInfo, poc, vah, val, tickSize float64) *StopLossInfo {
  entryPrice := result.EntryPrice
  currentPrice := result.CurrentPrice
  atr := result.ATR
  multiplier := result.ATRMultiplier

  atrStop, _ := decimal.NewFromFloat(entryPrice).Sub(
    decimal.NewFromFloat(atr).Mul(decimal.NewFromFloat(multiplier)),
  ).Float64()

  structureStop := r.VolumeProfileRepository.StructureSupport(entryPrice, poc, vah, val)
  if structureStop == 0 {
    structureStop = atrStop // 没有VP数据就用ATR
  } else {
    structureStop = structureStop * 0.998
  }

  percentStop, _ := decimal.NewFromFloat(entryPrice).Mul(
    decimal.NewFromFloat(0.95),
  ).Float64()

  limitStop := entryPrice * 0.998
  riskStop := entryPrice * (1 - result.Risk/float64(result.Leverage))

  result.InitialStop = r.roundToTick(
    math.Min(limitStop, math.Max(riskStop, math.Max(atrStop, math.Max(structureStop, percentStop)))),
    tickSize,
    "floor",
  )

  if atr > 0 {
    result.ProfitATR, _ = decimal.NewFromFloat(currentPrice - entryPrice).Div(
      decimal.NewFromFloat(atr),
    ).Round(2).Float64()
  }

  if result.ProfitATR >= 1.0 {
    result.BreakEvenStop = r.roundToTick(
      entryPrice+atr*0.1,
      tickSize,
      "ceil",
    )
  }

  if result.ProfitATR >= 2.5 {
    result.TrailingStop = r.roundToTick(currentPrice-atr*1.5, tickSize, "floor")
    result.StopType = "trailing_aggressive"
  } else if result.ProfitATR >= 1.5 {
    result.TrailingStop = result.BreakEvenStop
    result.StopType = "breakeven"
  } else {
    result.TrailingStop = result.InitialStop
    result.StopType = "initial"
  }

  target1 := entryPrice + atr*1.5
  if poc > 0 && poc > entryPrice {
    target1 = math.Max(target1, poc)
  }
  result.TakeProfit1 = r.roundToTick(target1, tickSize, "ceil")

  target2 := entryPrice + atr*3
  if vah > 0 && vah > entryPrice {
    target2 = math.Max(target2, vah)
  }
  result.TakeProfit2 = r.roundToTick(target2, tickSize, "ceil")

  risk := entryPrice - result.InitialStop
  reward := result.TakeProfit1 - entryPrice
  if risk > 0 {
    result.RiskReward, _ = decimal.NewFromFloat(reward).Div(
      decimal.NewFromFloat(risk),
    ).Round(2).Float64()
  }

  return result
}

func (r *StopLossRepository) calculateShortStops(result *StopLossInfo, poc, vah, val, tickSize float64) *StopLossInfo {
  entry := result.EntryPrice
  current := result.CurrentPrice
  atr := result.ATR
  multiplier := result.ATRMultiplier

  atrStop := entry + atr*multiplier

  structureStop := r.VolumeProfileRepository.StructureResistance(entry, poc, vah, val)
  if structureStop == 0 {
    structureStop = atrStop
  } else {
    structureStop = structureStop * 1.002 // 给结构位留 0.2% 的缓冲
  }

  percentStop := entry * 1.05

  limitStop := entry * 1.002
  riskStop := entry * (1 + result.Risk/float64(result.Leverage))

  result.InitialStop = r.roundToTick(
    math.Max(limitStop, math.Min(riskStop, math.Min(atrStop, math.Min(structureStop, percentStop)))),
    tickSize,
    "ceil",
  )

  if atr > 0 {
    result.ProfitATR, _ = decimal.NewFromFloat(entry - current).Div(
      decimal.NewFromFloat(atr),
    ).Round(2).Float64()
  }

  if result.ProfitATR >= 1.0 {
    result.BreakEvenStop = r.roundToTick(entry-atr*0.1, tickSize, "floor")
  }

  if result.ProfitATR >= 2.5 {
    result.TrailingStop = r.roundToTick(current+atr*1.5, tickSize, "ceil")
    result.StopType = "trailing_aggressive"
  } else if result.ProfitATR >= 1.5 {
    result.TrailingStop = result.BreakEvenStop
    result.StopType = "breakeven"
  } else {
    result.TrailingStop = result.InitialStop
    result.StopType = "initial"
  }

  target1 := entry - atr*1.5
  if poc > 0 && poc < entry {
    target1 = math.Min(target1, poc)
  }
  result.TakeProfit1 = r.roundToTick(target1, tickSize, "floor")

  target2 := entry - atr*3
  if val > 0 && val < entry {
    target2 = math.Min(target2, val)
  }
  result.TakeProfit2 = r.roundToTick(target2, tickSize, "floor")

  risk := result.InitialStop - entry
  reward := entry - result.TakeProfit1
  if risk > 0 {
    result.RiskReward, _ = decimal.NewFromFloat(reward).Div(
      decimal.NewFromFloat(risk),
    ).Round(2).Float64()
  }

  return result
}

func (r *StopLossRepository) updateLongTrailing(result *StopLossInfo, currentStop, tickSize float64) *StopLossInfo {
  newStop := result.TrailingStop
  if newStop > currentStop {
    result.ActiveStop = newStop
  } else {
    result.ActiveStop = currentStop
  }
  return result
}

func (r *StopLossRepository) updateShortTrailing(result *StopLossInfo, currentStop, tickSize float64) *StopLossInfo {
  newStop := result.TrailingStop
  if newStop < currentStop || currentStop == 0 {
    result.ActiveStop = newStop
  } else {
    result.ActiveStop = currentStop
  }
  return result
}

func (r *StopLossRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}

func (r *StopLossRepository) roundToTick(value, tickSize float64, direction string) float64 {
  if tickSize == 0 {
    return value
  }
  d := decimal.NewFromFloat(value).Div(decimal.NewFromFloat(tickSize))
  switch direction {
  case "ceil":
    d = d.Ceil()
  case "floor":
    d = d.Floor()
  default:
    d = d.Round(0)
  }
  result, _ := d.Mul(decimal.NewFromFloat(tickSize)).Float64()
  return result
}

func (r *StopLossRepository) ShouldTriggerStop(symbol string, side int, stopPrice float64) (result bool, err error) {
  price, err := r.TickersRepository.Price(symbol)
  if err != nil {
    return
  }
  if side == 1 { // LONG，价格跌破止损
    result = price <= stopPrice
    return
  }
  result = price >= stopPrice
  return
}

func (r *StopLossRepository) Save(symbol string, side int, result *StopLossInfo) error {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(config.REDIS_KEY_STOPLOSS, side, symbol, day)
  return r.Rdb.HMSet(r.Ctx, redisKey, map[string]interface{}{
    "side":          result.Side,
    "entry_price":   result.EntryPrice,
    "initial_stop":  result.InitialStop,
    "active_stop":   result.ActiveStop,
    "take_profit_1": result.TakeProfit1,
    "take_profit_2": result.TakeProfit2,
    "atr":           result.ATR,
    "stop_type":     result.StopType,
    "updated_at":    time.Now().Unix(),
  }).Err()
}
