package tradings

import (
  "errors"
  "fmt"
  "github.com/rs/xid"
  "log"
  "math"
  "time"

  "github.com/adshao/go-binance/v2/common"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  futuresModels "taoniu.local/cryptos/models/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures/tradings/triggers"
)

type TriggersRepository struct {
  Db                 *gorm.DB
  SymbolsRepository  SymbolsRepository
  PositionRepository PositionRepository
  OrdersRepository   OrdersRepository
}

func (r *TriggersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&futuresModels.Trigger{}).Where("status", []int{1, 3}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Ids() []string {
  var ids []string
  r.Db.Model(&futuresModels.Trigger{}).Select("id").Where("status", []int{1, 3}).Find(&ids)
  return ids
}

func (r *TriggersRepository) Place(id string) error {
  var trigger futuresModels.Trigger
  result := r.Db.First(&trigger, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("trigger empty")
  }

  if trigger.ExpiredAt.Unix() < time.Now().Unix() {
    trigger.Status = 4
    r.Db.Model(&futuresModels.Trigger{ID: trigger.ID}).Updates(trigger)
    return errors.New("trigger expired")
  }

  entity, err := r.SymbolsRepository.Get(trigger.Symbol)
  if err != nil {
    return err
  }

  tickSize, stepSize, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  position, err := r.PositionRepository.Get(trigger.Symbol, trigger.Side)
  if errors.Is(err, gorm.ErrRecordNotFound) && trigger.EntryQuantity > 0 {
    trigger.EntryQuantity = 0
    trigger.Profit, _ = decimal.NewFromFloat(trigger.Profit).Add(decimal.NewFromFloat(trigger.BufferProfit)).Float64()
    trigger.BufferProfit = 0
    r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Order("buy_price asc").Update("status", 4)
  }

  var positionSide string
  var side string
  if trigger.Side == 1 {
    positionSide = "LONG"
    side = "BUY"
  } else if trigger.Side == 2 {
    positionSide = "SHORT"
    side = "SELL"
  }

  if err == nil && position.Timestamp > trigger.Timestamp {
    trigger.EntryPrice = position.EntryPrice
    if trigger.EntryQuantity > 0 && position.Volume >= 0 {
      trigger.EntryQuantity = position.Volume
    }
    if trigger.Side == 2 && position.Volume <= 0 {
      trigger.EntryQuantity = -position.Volume
    }
    if trigger.EntryQuantity > 0 {
      trigger.TakePrice = r.TakePrice(trigger.EntryPrice, trigger.Side, tickSize)
      stopPrice, err := r.StopPrice(
        trigger.Capital,
        trigger.Side,
        position.Leverage,
        trigger.EntryPrice,
        trigger.EntryQuantity,
        tickSize,
        stepSize,
      )
      if err == nil {
        trigger.StopPrice = stopPrice
      }
    }
    trigger.Timestamp = position.Timestamp
    r.Db.Model(&futuresModels.Trigger{ID: trigger.ID}).Updates(trigger)
  }

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return err
  }

  if trigger.EntryQuantity > 0 {
    if trigger.Side == 1 {
      trigger.BufferProfit, _ = decimal.NewFromFloat(price).Sub(decimal.NewFromFloat(trigger.EntryPrice)).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Float64()
    } else {
      trigger.BufferProfit, _ = decimal.NewFromFloat(trigger.EntryPrice).Sub(decimal.NewFromFloat(price)).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Float64()
    }
    r.Db.Model(&futuresModels.Trigger{ID: trigger.ID}).UpdateColumn("buffer_profit", trigger.BufferProfit)
  }

  var buyPrice float64
  var buyQuantity float64

  if trigger.EntryPrice == 0 {
    buyPrice = trigger.Price
    buyQuantity, _ = decimal.NewFromFloat(5).Div(decimal.NewFromFloat(trigger.Price)).Float64()
  }

  if trigger.EntryPrice > 0 && trigger.EntryQuantity == 0 {
    buyQuantity = r.BuyQuantity(trigger.Side, 5, trigger.EntryPrice, 5)
    buyPrice, _ = decimal.NewFromFloat(5).Div(decimal.NewFromFloat(buyQuantity)).Float64()
  }

  entryAmount, _ := decimal.NewFromFloat(trigger.EntryPrice).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Float64()

  ipart, _ := math.Modf(trigger.Capital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  if buyPrice == 0 {
    capital, err := r.Capital(trigger.Capital, entryAmount, places)
    if err != nil {
      return errors.New("reach the max invest capital")
    }
    ratio := r.Ratio(capital, entryAmount)
    buyAmount, _ := decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if buyAmount < 5 {
      buyAmount = 5
    }
    buyQuantity = r.BuyQuantity(trigger.Side, buyAmount, trigger.EntryPrice, entryAmount)
    buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
  }

  if trigger.Side == 1 {
    if price < buyPrice {
      buyPrice = price
    }
    buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    if price > buyPrice {
      buyPrice = price
    }
    buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if trigger.Side == 1 && price > buyPrice {
    return errors.New(fmt.Sprintf("price must reach %v", buyPrice))
  }
  if trigger.Side == 2 && price < buyPrice {
    return errors.New(fmt.Sprintf("price must reach %v", buyPrice))
  }

  buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
  trigger.EntryQuantity, _ = decimal.NewFromFloat(trigger.EntryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()

  buyAmount, _ := decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
  entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()

  if trigger.EntryPrice == 0 {
    trigger.EntryPrice = buyPrice
  } else {
    trigger.EntryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(trigger.EntryQuantity)).Float64()
  }

  sellPrice := r.SellPrice(trigger.Side, trigger.EntryPrice, entryAmount)
  if trigger.Side == 1 {
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if !r.CanBuy(&trigger, buyPrice) {
    return errors.New("can not buy now")
  }

  r.Db.Transaction(func(tx *gorm.DB) error {
    orderID, err := r.OrdersRepository.Create(trigger.Symbol, positionSide, side, buyPrice, buyQuantity)
    if err != nil {
      apiError, ok := err.(common.APIError)
      if ok {
        if apiError.Code == -2010 {
          return err
        }
      }
      trigger.Remark = err.Error()
    }
    trigger.Timestamp = 0
    if err := tx.Model(&futuresModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
      return err
    }
    grid := models.Grid{
      ID:           xid.New().String(),
      Symbol:       trigger.Symbol,
      TriggerID:    trigger.ID,
      BuyOrderId:   orderID,
      BuyPrice:     buyPrice,
      BuyQuantity:  buyQuantity,
      SellPrice:    sellPrice,
      SellQuantity: buyQuantity,
      Status:       0,
    }
    if err := tx.Create(&grid).Error; err != nil {
      return err
    }

    return nil
  })

  return nil
}

func (r *TriggersRepository) Flush(id string) error {
  var trigger futuresModels.Trigger
  result := r.Db.First(&trigger, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("fishers empty")
  }

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(&trigger, price)
  if err != nil {
    log.Println("take error", err)
  }

  var positionSide string
  if trigger.Side == 1 {
    positionSide = "LONG"
  } else if trigger.Side == 2 {
    positionSide = "SHORT"
  }

  var grids []*models.Grid
  r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 2}).Find(&grids)
  for _, grid := range grids {
    if grid.Status == 0 {
      timestamp := grid.CreatedAt.Unix()
      if grid.BuyOrderId == 0 {
        var side string
        if trigger.Side == 1 {
          side = "BUY"
        } else if trigger.Side == 2 {
          side = "SELL"
        }
        orderID := r.OrdersRepository.Lost(trigger.Symbol, positionSide, side, grid.BuyPrice, timestamp-30)
        if orderID > 0 {
          grid.BuyOrderId = orderID
          if err := r.Db.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
            return err
          }
        }
      } else {
        if timestamp < time.Now().Unix()-300 {
          r.OrdersRepository.Cancel(grid.Symbol, grid.BuyOrderId)
        }
      }

      status := r.OrdersRepository.Status(grid.Symbol, grid.BuyOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(grid.Symbol, grid.BuyOrderId)
        continue
      }

      r.Db.Transaction(func(tx *gorm.DB) error {
        if status == "FILLED" {
          grid.Status = 1
        } else {
          trigger.Timestamp = 0
          if err := tx.Model(&futuresModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
            return err
          }
          grid.Status = 4
        }
        if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
          return err
        }
        return nil
      })
    } else if grid.Status == 2 {
      timestamp := grid.UpdatedAt.Unix()
      if grid.SellOrderId == 0 {
        var side string
        if trigger.Side == 1 {
          side = "SELL"
        } else if trigger.Side == 2 {
          side = "BUY"
        }
        orderID := r.OrdersRepository.Lost(trigger.Symbol, positionSide, side, grid.SellPrice, timestamp-30)
        if orderID > 0 {
          grid.SellOrderId = orderID
          if err := r.Db.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
            return err
          }
        } else {
          if timestamp < time.Now().Unix()-300 {
            r.Db.Model(&models.Grid{ID: grid.ID}).Update("status", 1)
          }
          return nil
        }
      }

      status := r.OrdersRepository.Status(trigger.Symbol, grid.SellOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(trigger.Symbol, grid.SellOrderId)
        continue
      }

      r.Db.Transaction(func(tx *gorm.DB) error {
        if status == "FILLED" {
          grid.Status = 3
        } else {
          trigger.Timestamp = 0
          if err := tx.Model(&futuresModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
            return err
          }
          grid.Status = 5
        }
        if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
          return err
        }
        return nil
      })
    }
  }

  return nil
}

func (r *TriggersRepository) Take(trigger *futuresModels.Trigger, price float64) error {
  var positionSide string
  var side string
  if trigger.Side == 1 {
    positionSide = "LONG"
    side = "SELL"
  } else if trigger.Side == 2 {
    positionSide = "SHORT"
    side = "BUY"
  }

  var grid models.Grid
  if trigger.Side == 1 {
    result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Order("sell_price asc").Take(&grid)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("empty grid")
    }
    if price < grid.SellPrice*0.995 {
      if price < trigger.EntryPrice*1.035 {
        return errors.New("price too low")
      }
      grid.SellPrice = price
    }
  }
  if trigger.Side == 2 && price > grid.SellPrice {
    result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Order("sell_price desc").Take(&grid)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("empty grid")
    }
    if price > grid.SellPrice*1.005 {
      if price > trigger.EntryPrice*0.965 {
        return errors.New("price too high")
      }
      grid.SellPrice = price
    }
  }
  r.Db.Transaction(func(tx *gorm.DB) error {
    trigger.EntryQuantity, _ = decimal.NewFromFloat(trigger.EntryQuantity).Sub(decimal.NewFromFloat(grid.SellQuantity)).Float64()
    orderID, err := r.OrdersRepository.Create(grid.Symbol, positionSide, side, grid.SellPrice, grid.SellQuantity)
    if err != nil {
      apiError, ok := err.(common.APIError)
      if ok {
        if apiError.Code == -2010 {
          tx.Model(&futuresModels.Trigger{ID: trigger.ID}).Update("remark", err.Error())
          return nil
        }
      }
      trigger.Remark = err.Error()
    }
    if err := tx.Model(&futuresModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
      return err
    }
    grid.SellOrderId = orderID
    grid.Status = 2
    if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
      return err
    }
    return nil
  })
  return nil
}

func (r *TriggersRepository) Ratio(capital float64, entryAmount float64) float64 {
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
  return 0.0
}

func (r *TriggersRepository) Capital(capital float64, entryAmount float64, place int) (result float64, err error) {
  step := math.Pow10(place - 1)

  for {
    ratio := r.Ratio(capital, entryAmount)
    if ratio == 0 {
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

func (r *TriggersRepository) BuyQuantity(
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

func (r *TriggersRepository) SellPrice(
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

func (r *TriggersRepository) TakePrice(
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

func (r *TriggersRepository) StopPrice(
  maxCapital float64,
  side int,
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
    buyQuantity = r.BuyQuantity(side, buyAmount, entryPrice, entryAmount)
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

func (r *TriggersRepository) CanBuy(
  trigger *futuresModels.Trigger,
  price float64,
) bool {
  var grid models.Grid
  if trigger.Side == 1 {
    result := r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Order("buy_price asc").Take(&grid)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if price >= grid.BuyPrice {
        return false
      }
    }
  }
  if trigger.Side == 2 {
    result := r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Order("buy_price desc").Take(&grid)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if price <= grid.BuyPrice {
        return false
      }
    }
  }

  return true
}
