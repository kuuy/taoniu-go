package futures

import (
  "errors"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"
  "math"
  models "taoniu.local/cryptos/models/binance/futures"
)

type PositionsRepository struct {
  Db *gorm.DB
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
    lost, _ = decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.005)).Float64()
    if side == 1 {
      entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.995)).Float64()
      buyQuantity, _ = decimal.NewFromFloat(buyAmount).Add(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
    } else {
      entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.005)).Float64()
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
      sellPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.005)).Float64()
    } else {
      sellPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.995)).Float64()
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
