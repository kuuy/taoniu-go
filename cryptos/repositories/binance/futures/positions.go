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
    if entryAmount == 0.0 {
      return ratio
    }
    if totalAmount >= entryAmount-lastAmount {
      return ratio
    }
    lastAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    totalAmount, _ = decimal.NewFromFloat(totalAmount).Add(decimal.NewFromFloat(lastAmount)).Float64()
  }
  return 0.0
}

func (r *PositionsRepository) Calc(
  capital float64,
  side int,
  entryPrice float64,
  entryAmount float64,
  ratio float64,
) (float64, float64, float64) {
  lost, _ := decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.005)).Float64()
  amount, _ := decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()

  var volume float64
  if side == 1 {
    volume, _ = decimal.NewFromFloat(amount).Add(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
  } else {
    volume, _ = decimal.NewFromFloat(amount).Sub(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
  }

  price, _ := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(volume)).Float64()

  return price, volume, amount
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

  if result == 0.0 {
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

  return
}
