package futures

import (
  "errors"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"
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
  //ratios := []float64{0.0071, 0.0216, 0.0361, 0.0586, 0.0739, 0.1122}
  //ratios := []float64{1,           0.328703704, 0.59833795, 0.616040956, 0.792963464, 0.658645276}
  //ratios := []float64{3.042253521, 1.671296296, 1.623268698, 1.26109215, 1.51826793, 1}
  //1.718309859 1.029586967 0.971263265
  //ratios := []float64{0.0071, 0.0122, 0.0209, 0.0358, 0.0614, 0.1053}
  ratios := []float64{0.0071, 0.0193, 0.0331, 0.0567, 0.0972, 0.1667}
  //ratios := []float64{0.0071, 0.0126, 0.0222, 0.0391, 0.0690, 0.1218}
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
