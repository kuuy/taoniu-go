package tradings

import (
  "errors"
  "fmt"
  "log"
  "time"

  "github.com/adshao/go-binance/v2/common"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  crossModels "taoniu.local/cryptos/models/binance/spot/margin/cross"
  models "taoniu.local/cryptos/models/binance/spot/margin/cross/tradings/triggers"
)

type TriggersRepository struct {
  Db                      *gorm.DB
  SymbolsRepository       SymbolsRepository
  AccountRepository       AccountRepository
  MarginAccountRepository MarginAccountRepository
  OrdersRepository        OrdersRepository
}

func (r *TriggersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&crossModels.Trigger{}).Where("status", []int{1, 3}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Place(symbol string) error {
  var trigger crossModels.Trigger
  result := r.Db.Where("symbol=? AND status=?", symbol, 1).Take(&trigger)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("trigger empty")
  }

  if trigger.ExpiredAt.Unix() < time.Now().Unix() {
    trigger.Status = 4
    r.Db.Model(&crossModels.Trigger{ID: trigger.ID}).Updates(trigger)
    return errors.New("trigger expired")
  }

  entity, err := r.SymbolsRepository.Get(symbol)
  if err != nil {
    return err
  }

  tickSize, stepSize, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  entryAmount, _ := decimal.NewFromFloat(trigger.EntryPrice).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Float64()
  ratio := r.Ratio(trigger.Capital, entryAmount)
  if ratio == 0.0 {
    return errors.New("reach the max invest capital")
  }
  buyPrice, buyQuantity := r.Calc(trigger.Capital, 1, trigger.EntryPrice, entryAmount, ratio)
  buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
  buyAmount, _ := decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()

  entryQuantity, _ := decimal.NewFromFloat(trigger.EntryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()

  price, err := r.SymbolsRepository.Price(symbol)
  if err != nil {
    return err
  }

  if price > buyPrice {
    return errors.New(fmt.Sprintf("price must reach %v", buyPrice))
  }

  if !r.CanBuy(symbol, buyPrice) {
    return errors.New("can not buy now")
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return err
  }
  if balance["free"] < buyAmount {
    transferId, err := r.MarginAccountRepository.Loan(entity.QuoteAsset, "", buyAmount, false)
    if err != nil {
      return err
    }
    log.Println("loan", transferId, buyAmount)
  }

  if trigger.EntryQuantity == 0.0 {
    trigger.EntryPrice = buyPrice
  } else {
    trigger.EntryPrice, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Div(decimal.NewFromFloat(entryQuantity)).Float64()
  }
  trigger.EntryQuantity = entryQuantity

  sellPrice, _ := decimal.NewFromFloat(trigger.EntryPrice).Mul(decimal.NewFromFloat(1.02)).Float64()
  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  sellQuantity, _ := decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(sellPrice)).Float64()
  sellQuantity, _ = decimal.NewFromFloat(sellQuantity).Div(decimal.NewFromFloat(stepSize)).Floor().Mul(decimal.NewFromFloat(stepSize)).Float64()

  r.Db.Transaction(func(tx *gorm.DB) error {
    orderID, err := r.OrdersRepository.Create(symbol, "BUY", buyPrice, buyQuantity, false)
    if err != nil {
      apiError, ok := err.(common.APIError)
      if ok {
        if apiError.Code == -2010 {
          return err
        }
      }
      trigger.Remark = err.Error()
    }
    if err := tx.Model(&crossModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
      return err
    }
    grid := models.Grid{
      ID:           xid.New().String(),
      Symbol:       symbol,
      TriggerID:    trigger.ID,
      BuyOrderId:   orderID,
      BuyPrice:     buyPrice,
      BuyQuantity:  buyQuantity,
      SellPrice:    sellPrice,
      SellQuantity: sellQuantity,
      Status:       0,
    }
    if err := tx.Create(&grid).Error; err != nil {
      return err
    }

    return nil
  })

  return nil
}

func (r *TriggersRepository) Flush(symbol string) error {
  var trigger crossModels.Trigger
  result := r.Db.Where("symbol=?", symbol).Take(&trigger)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("fishers empty")
  }

  price, err := r.SymbolsRepository.Price(symbol)
  if err != nil {
    return err
  }
  err = r.Take(&trigger, price)
  if err != nil {
    log.Println("take error", err)
  }

  var grids []*models.Grid
  r.Db.Where("symbol=? AND status IN ?", trigger.Symbol, []int{0, 2}).Find(&grids)
  for _, grid := range grids {
    if grid.Status == 0 {
      timestamp := grid.CreatedAt.Unix()
      if grid.BuyOrderId == 0 {
        orderID := r.OrdersRepository.Lost(symbol, "BUY", grid.BuyPrice, timestamp-30)
        if orderID > 0 {
          grid.BuyOrderId = orderID
          if err := r.Db.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
            return err
          }
        } else {
          if timestamp > time.Now().Unix()-300 {
            r.Db.Transaction(func(tx *gorm.DB) error {
              buyAmount := decimal.NewFromFloat(grid.BuyPrice).Mul(decimal.NewFromFloat(grid.BuyQuantity))
              entryAmount, _ := decimal.NewFromFloat(trigger.EntryPrice).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Sub(buyAmount).Float64()
              entryQuantity, _ := decimal.NewFromFloat(trigger.EntryQuantity).Sub(decimal.NewFromFloat(grid.BuyQuantity)).Float64()
              if entryQuantity == 0.0 {
                trigger.EntryPrice = trigger.Price
              } else {
                trigger.EntryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
              }
              trigger.EntryQuantity = entryQuantity
              if err := tx.Model(&crossModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
                return err
              }
              grid.Status = 4
              if err := tx.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
                return err
              }
              return nil
            })
          }
          return nil
        }
      }
      status := r.OrdersRepository.Status(symbol, grid.BuyOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(symbol, grid.BuyOrderId, false)
        continue
      }
      r.Db.Transaction(func(tx *gorm.DB) error {
        if status == "FILLED" {
          grid.Status = 1
        } else {
          buyAmount := decimal.NewFromFloat(grid.BuyPrice).Mul(decimal.NewFromFloat(grid.BuyQuantity))
          entryAmount, _ := decimal.NewFromFloat(trigger.EntryPrice).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Sub(buyAmount).Float64()
          entryQuantity, _ := decimal.NewFromFloat(trigger.EntryQuantity).Sub(decimal.NewFromFloat(grid.BuyQuantity)).Float64()
          if entryQuantity == 0.0 {
            trigger.EntryPrice = trigger.Price
          } else {
            trigger.EntryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
          }
          trigger.EntryQuantity = entryQuantity
          if err := tx.Model(&crossModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
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
        orderID := r.OrdersRepository.Lost(symbol, "SELL", grid.SellPrice, timestamp-30)
        if orderID > 0 {
          grid.SellOrderId = orderID
          if err := r.Db.Model(&models.Grid{ID: grid.ID}).Updates(grid).Error; err != nil {
            return err
          }
        } else {
          if timestamp > time.Now().Unix()-300 {
            r.Db.Model(&models.Grid{ID: grid.ID}).Update("status", 1)
          }
          return nil
        }
      }
      status := r.OrdersRepository.Status(symbol, grid.SellOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(symbol, grid.SellOrderId, false)
        continue
      }
      r.Db.Transaction(func(tx *gorm.DB) error {
        if status == "FILLED" {
          grid.Status = 3
        } else {
          buyAmount := decimal.NewFromFloat(grid.BuyPrice).Mul(decimal.NewFromFloat(grid.BuyQuantity))
          entryAmount, _ := decimal.NewFromFloat(trigger.EntryPrice).Mul(decimal.NewFromFloat(trigger.EntryQuantity)).Sub(buyAmount).Float64()
          entryQuantity, _ := decimal.NewFromFloat(trigger.EntryQuantity).Sub(decimal.NewFromFloat(grid.BuyQuantity)).Float64()
          if entryQuantity == 0.0 {
            trigger.EntryPrice = trigger.Price
          } else {
            trigger.EntryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
          }
          trigger.EntryQuantity = entryQuantity
          if err := tx.Model(&crossModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
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

func (r *TriggersRepository) Take(trigger *crossModels.Trigger, price float64) error {
  var grid models.Grid
  result := r.Db.Where("symbol=? AND status=?", trigger.Symbol, 1).Order("sell_price asc").Take(&grid)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty grid")
  }
  if price < grid.SellPrice {
    return errors.New("price too low")
  }
  r.Db.Transaction(func(tx *gorm.DB) error {
    trigger.EntryQuantity, _ = decimal.NewFromFloat(trigger.EntryQuantity).Sub(decimal.NewFromFloat(grid.SellQuantity)).Float64()
    orderID, err := r.OrdersRepository.Create(grid.Symbol, "SELL", grid.SellPrice, grid.SellQuantity, false)
    if err != nil {
      apiError, ok := err.(common.APIError)
      if ok {
        if apiError.Code == -2010 {
          tx.Model(&crossModels.Trigger{ID: trigger.ID}).Update("remark", err.Error())
          return nil
        }
      }
      trigger.Remark = err.Error()
    }
    if err := tx.Model(&crossModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
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

func (r *TriggersRepository) Calc(
  capital float64,
  side int,
  entryPrice float64,
  entryAmount float64,
  ratio float64,
) (float64, float64) {
  lost, _ := decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.005)).Float64()
  amount, _ := decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()

  var quantity float64
  if side == 1 {
    quantity, _ = decimal.NewFromFloat(amount).Add(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
  } else {
    quantity, _ = decimal.NewFromFloat(amount).Sub(decimal.NewFromFloat(lost)).Div(decimal.NewFromFloat(entryPrice)).Float64()
  }

  price, _ := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(quantity)).Float64()

  return price, quantity
}

func (r *TriggersRepository) CanBuy(
  symbol string,
  price float64,
) bool {
  var grid models.Grid
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{0, 1, 2}).Order("buy_price asc").Take(&grid)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if price >= grid.BuyPrice {
      return false
    }
  }

  return true
}
