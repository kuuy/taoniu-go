package futures

import (
  "errors"
  "log"
  "math"

  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type PositionsRepository struct {
  Db                *gorm.DB
  SymbolsRepository *SymbolsRepository
}

func (r *PositionsRepository) Get(
  symbol string,
  side int,
) (models.Position, error) {
  var entity models.Position
  result := r.Db.Where("symbol=? AND side=? AND status=1", symbol, side).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return entity, result.Error
  }
  return entity, nil
}

func (r *PositionsRepository) Gets(conditions map[string]interface{}) []*models.Position {
  var positions []*models.Position
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "leverage",
    "capital",
    "notional",
    "entry_price",
    "entry_quantity",
    "timestamp",
  })
  if _, ok := conditions["side"]; ok {
    query.Where("side", conditions["side"].(int))
  }
  query.Where("status=1 AND entry_quantity>0").Find(&positions)
  return positions
}

func (r *PositionsRepository) Ratio(capital float64, entryAmount float64) float64 {
  totalAmount := 0.0
  lastAmount := 0.0
  ratios := []float64{0.0071, 0.0193, 0.0331, 0.0567, 0.0972, 0.1667}
  for _, ratio := range ratios {
    if entryAmount == 0 {
      return ratio
    }
    if totalAmount >= entryAmount-lastAmount {
      return ratio
    }
    lastAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    totalAmount, _ = decimal.NewFromFloat(totalAmount).Add(decimal.NewFromFloat(lastAmount)).Float64()
  }
  return 0
}

func (r *PositionsRepository) BuyQuantity(
  side int,
  buyAmount float64,
  entryPrice float64,
  entryAmount float64,
) (buyQuantity float64) {
  ipart, _ := math.Modf(entryAmount + buyAmount)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }
  var lost float64
  for i := 0; i < places; i++ {
    lost, _ = decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.0085)).Float64()
    if side == 1 {
      entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.9915)).Float64()
      buyQuantity, _ = decimal.NewFromFloat(buyAmount).Add(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
    } else {
      entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.0085)).Float64()
      buyQuantity, _ = decimal.NewFromFloat(buyAmount).Sub(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
    }
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(lost)).Float64()
  }
  return
}

func (r *PositionsRepository) SellPrice(
  side int,
  entryPrice float64,
  entryAmount float64,
) (sellPrice float64) {
  ipart, _ := math.Modf(entryAmount)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }
  for i := 0; i < places; i++ {
    if side == 1 {
      sellPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.0085)).Float64()
    } else {
      sellPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.9915)).Float64()
    }
  }
  return
}

func (r *PositionsRepository) TakePrice(
  entryPrice float64,
  side int,
  tickSize float64,
) float64 {
  var takePrice float64
  if side == 1 {
    takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.02)).Float64()
    takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.98)).Float64()
    takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }
  return takePrice
}

func (r *PositionsRepository) StopPrice(
  maxCapital float64,
  side int,
  price float64,
  leverage int,
  entryPrice float64,
  entryQuantity float64,
  tickSize float64,
  stepSize float64,
) (stopPrice float64, err error) {
  ipart, _ := math.Modf(maxCapital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64

  for {
    var err error
    capital, err := r.Capital(maxCapital, entryAmount, places)
    if err != nil {
      break
    }
    ratio := r.Ratio(capital, entryAmount)
    buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if buyAmount < 5 {
      buyAmount = 5
    }

    if entryAmount == 0 {
      buyAmount = 5
      buyQuantity, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(price)).Float64()
    } else {
      buyQuantity = r.BuyQuantity(side, buyAmount, entryPrice, entryAmount)
    }

    buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
    if side == 1 {
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
  }

  stopAmount, _ := decimal.NewFromFloat(entryAmount).Div(decimal.NewFromInt32(int32(leverage))).Mul(decimal.NewFromFloat(0.1)).Float64()
  if side == 1 {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Sub(
      decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
    ).Float64()
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Add(
      decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
    ).Float64()
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  return
}

func (r *PositionsRepository) Capital(capital float64, entryAmount float64, place int) (result float64, err error) {
  step := math.Pow10(place - 1)

  for {
    ratio := r.Ratio(capital, entryAmount)
    if ratio == 0.0 {
      break
    }
    result = capital
    if capital <= step {
      break
    }
    capital -= step
  }

  if result == 0 {
    err = errors.New("reach the max invest capital")
    return
  }

  if place > 1 {
    capital, err = r.Capital(result+step, entryAmount, place-1)
    if err != nil {
      return
    }
    result = capital
  }

  if result < 5 {
    result = 5
  }

  return
}

func (r *PositionsRepository) Flush(symbol string, side int) error {
  var updateTime int64
  updateTime = 1693219638260

  var positionSide string
  if side == 1 {
    positionSide = "LONG"
  } else if side == 2 {
    positionSide = "SHORT"
  }
  var orders []*models.Order
  r.Db.Model(models.Order{}).Select([]string{
    "symbol",
    "position_side",
    "avg_price",
    "executed_quantity",
    "side",
    "update_time",
  }).Where("symbol=? AND position_side=? AND update_time>?", symbol, positionSide, updateTime).Order("update_time asc").Find(&orders)
  var entryAmount float64
  var entryPrice float64
  var entryQuantity float64
  for _, order := range orders {
    updateTime = order.UpdateTime
    if order.Status == "PARTIALLY_FILLED" || order.ExecutedQuantity == 0 {
      continue
    }
    executedQuantity := decimal.NewFromFloat(order.ExecutedQuantity)
    executedAmount := decimal.NewFromFloat(order.AvgPrice).Mul(executedQuantity)
    if side == 1 && order.Side == "BUY" || side == 2 && order.Side == "SELL" {
      entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(executedAmount).Float64()
      entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(executedQuantity).Float64()
      entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
    }
    if side == 1 && order.Side == "SELL" || side == 2 && order.Side == "BUY" {
      entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Sub(executedQuantity).Float64()
      entryAmount, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()
    }
    log.Println("order", symbol, side, order.Side, order.AvgPrice, order.ExecutedQuantity, order.UpdateTime, entryQuantity)
  }
  log.Println("position flush", symbol, side, entryAmount, entryPrice, entryQuantity, updateTime)
  return nil
}

func (r *PositionsRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  return r.SymbolsRepository.Filters(entity.Filters)
}
