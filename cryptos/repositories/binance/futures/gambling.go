package futures

import (
  "math"

  "github.com/shopspring/decimal"
  "gorm.io/gorm"
)

type GamblingRepository struct {
  Db                *gorm.DB
  SymbolsRepository *SymbolsRepository
}

func (r *GamblingRepository) Factors(entryAmount float64, side int) (factors [][]float64) {
  if entryAmount < 2000 {
    if side == 1 {
      factors = append(factors, []float64{1.0105, 0.25})
    } else {
      factors = append(factors, []float64{0.9895, 0.25})
    }
  } else {
    if side == 1 {
      factors = append(factors, []float64{1.0085, 0.25})
      factors = append(factors, []float64{1.0105, 0.5})
    } else {
      factors = append(factors, []float64{0.9915, 0.25})
      factors = append(factors, []float64{0.9895, 0.5})
    }
  }
  return
}

func (r *GamblingRepository) BuyQuantity(
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

func (r *GamblingRepository) SellPrice(
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

func (r *GamblingRepository) TakePrice(
  entryPrice float64,
  side int,
  tickSize float64,
) float64 {
  var takePrice float64
  if side == 1 {
    takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.0344)).Float64()
    takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.9656)).Float64()
    takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }
  return takePrice
}

func (r *GamblingRepository) StopPrice(
  entryPrice float64,
  side int,
  tickSize float64,
) float64 {
  var stopPrice float64
  if side == 1 {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.9828)).Float64()
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.0172)).Float64()
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }
  return stopPrice
}

func (r *GamblingRepository) Calc(
  entryPrice float64,
  entryQuantity float64,
  side int,
  tickSize float64,
  stepSize float64,
) (plans []*GamblingPlan) {
  var takePrice float64
  var takeQuantity float64
  var takeAmount float64
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()
  for _, factor := range r.Factors(entryAmount, side) {
    takeQuantity, _ = decimal.NewFromFloat(entryQuantity).Mul(decimal.NewFromFloat(factor[1])).Float64()
    takeQuantity, _ = decimal.NewFromFloat(takeQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    if side == 1 {
      takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(factor[0])).Float64()
      takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(factor[0])).Float64()
      takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    if entryQuantity <= takeQuantity {
      break
    }
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Sub(decimal.NewFromFloat(takeQuantity)).Float64()
    takeAmount, _ = decimal.NewFromFloat(takePrice).Mul(decimal.NewFromFloat(takeQuantity)).Float64()
    plans = append(plans, &GamblingPlan{
      TakePrice:    takePrice,
      TakeQuantity: takeQuantity,
      TakeAmount:   takeAmount,
    })
  }
  return
}

func (r *GamblingRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }
  tickSize, stepSize, _, err = r.SymbolsRepository.Filters(entity.Filters)
  return
}
