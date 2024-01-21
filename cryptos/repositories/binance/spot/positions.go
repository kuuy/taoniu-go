package spot

import (
  "errors"
  "github.com/rs/xid"
  "math"
  "time"

  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot"
)

type PositionsRepository struct {
  Db                *gorm.DB
  SymbolsRepository *SymbolsRepository
}

func (r *PositionsRepository) Get(symbol string) (position *models.Position, err error) {
  err = r.Db.Where("symbol=? AND status=1", symbol).Take(&position).Error
  return
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
    entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.9915)).Float64()
    buyQuantity, _ = decimal.NewFromFloat(buyAmount).Add(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(lost)).Float64()
  }
  return
}

func (r *PositionsRepository) SellQuantity(
  sellAmount float64,
  entryPrice float64,
  entryAmount float64,
) (sellQuantity float64) {
  ipart, _ := math.Modf(entryAmount - sellAmount)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }
  var lost float64
  for i := 0; i < places; i++ {
    lost, _ = decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.0085)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.0085)).Float64()
    sellQuantity, _ = decimal.NewFromFloat(sellAmount).Sub(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(lost)).Float64()
  }
  return
}

func (r *PositionsRepository) SellPrice(
  entryPrice float64,
  entryAmount float64,
) (sellPrice float64) {
  ipart, _ := math.Modf(entryAmount)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }
  for i := 0; i < places; i++ {
    sellPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.0085)).Float64()
  }
  return
}

func (r *PositionsRepository) TakePrice(
  entryPrice float64,
  tickSize float64,
) float64 {
  var takePrice float64
  takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.02)).Float64()
  takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  return takePrice
}

func (r *PositionsRepository) StopPrice(
  maxCapital float64,
  price float64,
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
      buyQuantity = r.BuyQuantity(buyAmount, entryPrice, entryAmount)
    }

    buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
    buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
  }

  stopAmount, _ := decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.1)).Float64()
  stopPrice, _ = decimal.NewFromFloat(entryPrice).Sub(
    decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
  ).Float64()
  stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

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

func (r *PositionsRepository) Apply(symbol string) error {
  timestamp := time.Now().UnixMilli() - 300*1e3
  var entity *models.Position
  result := r.Db.Where("symbol", symbol).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = &models.Position{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Timestamp: timestamp,
      Status:    1,
    }
    r.Db.Create(&entity)
  } else {
    if entity.EntryQuantity == 0 {
      r.Db.Model(&entity).Updates(map[string]interface{}{
        "timestamp": timestamp,
        "status":    1,
      })
    }
  }
  return nil
}

func (r *PositionsRepository) Flush(position *models.Position) (err error) {
  var orders []*models.Order
  r.Db.Model(models.Order{}).Select([]string{
    "symbol",
    "price",
    "executed_quantity",
    "side",
    "update_time",
    "status",
  }).Where(
    "symbol=? AND update_time>? AND status IN ?",
    position.Symbol,
    position.Timestamp,
    []string{"NEW", "PARTIALLY_FILLED", "FILLED"},
  ).Order("update_time asc").Find(&orders)
  if len(orders) == 0 {
    return
  }
  var entryPrice float64
  var entryAmount float64
  var entryQuantity float64
  for _, order := range orders {
    if order.Status != "FILLED" || order.ExecutedQuantity < order.Quantity {
      continue
    }
    executedQuantity := decimal.NewFromFloat(order.ExecutedQuantity)
    executedAmount := decimal.NewFromFloat(order.Price).Mul(executedQuantity)
    if order.Side == "BUY" {
      entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(executedAmount).Float64()
      entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(executedQuantity).Float64()
      entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
    }
    if order.Side == "SELL" {
      entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Sub(executedQuantity).Float64()
      if entryQuantity <= 0 {
        entryPrice = 0
        entryQuantity = 0
        entryAmount = 0
      } else {
        entryAmount, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()
      }
    }
    position.Timestamp = order.UpdateTime
  }
  values := map[string]interface{}{
    "entry_price":    entryPrice,
    "entry_quantity": entryQuantity,
    "entry_amount":   entryAmount,
    "version":        gorm.Expr("version + ?", 1),
  }
  if entryQuantity == 0 {
    values["timestamp"] = position.Timestamp
  }
  r.Db.Model(&position).Where("version", position.Version).Updates(values)
  return
}

func (r *PositionsRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}
