package tradings

import (
  "errors"
  "fmt"
  "log"
  "math"
  "time"

  "github.com/adshao/go-binance/v2/common"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  futuresModels "taoniu.local/cryptos/models/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures/tradings"
)

type TriggersRepository struct {
  Db                 *gorm.DB
  SymbolsRepository  SymbolsRepository
  AccountRepository  AccountRepository
  OrdersRepository   OrdersRepository
  PositionRepository PositionRepository
}

func (r *TriggersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&futuresModels.Trigger{}).Where("status", 1).Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Ids() []string {
  var ids []string
  r.Db.Model(&futuresModels.Trigger{}).Where("status", 1).Pluck("id", &ids)
  return ids
}

func (r *TriggersRepository) TriggerIds() []string {
  var ids []string
  r.Db.Model(&models.Trigger{}).Select("trigger_id").Where("status", []int{0, 1, 2}).Distinct("trigger_id").Find(&ids)
  return ids
}

func (r *TriggersRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Trigger{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["status"]; ok {
    query.Where("status IN ?", conditions["status"].([]int))
  } else {
    query.Where("status IN ?", []int{0, 1, 2, 3})
  }
  query.Count(&total)
  return total
}

func (r *TriggersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Trigger {
  var tradings []*models.Trigger
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "trigger_id",
    "buy_price",
    "buy_quantity",
    "sell_price",
    "sell_quantity",
    "buy_order_id",
    "sell_order_id",
    "status",
    "created_at",
    "updated_at",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["status"]; ok {
    query.Where("status IN ?", conditions["status"].([]int))
  } else {
    query.Where("status IN ?", []int{0, 1, 2, 3})
  }
  query.Order("updated_at desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&tradings)
  return tradings
}

func (r *TriggersRepository) Place(id string) error {
  var trigger *futuresModels.Trigger
  result := r.Db.First(&trigger, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("trigger empty")
  }

  if trigger.ExpiredAt.Unix() < time.Now().Unix() {
    r.Db.Model(&trigger).Update("status", 4)
    return errors.New("trigger expired")
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

  position, err := r.PositionRepository.Get(trigger.Symbol, trigger.Side)
  if err != nil {
    return err
  }

  if position.EntryQuantity == 0 {
    r.Close(trigger)
  }

  entity, err := r.SymbolsRepository.Get(trigger.Symbol)
  if err != nil {
    return err
  }

  tickSize, stepSize, notional, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  entryPrice := position.EntryPrice
  entryQuantity := position.EntryQuantity
  if position.Timestamp > trigger.Timestamp {
    trigger.Timestamp = position.Timestamp
    trigger.TakePrice = r.PositionRepository.TakePrice(entryPrice, trigger.Side, tickSize)
    stopPrice, err := r.PositionRepository.StopPrice(
      trigger.Capital,
      trigger.Side,
      trigger.Price,
      position.Leverage,
      entryPrice,
      entryQuantity,
      tickSize,
      stepSize,
    )
    if err == nil {
      trigger.StopPrice = stopPrice
    }

    err = r.Db.Model(&trigger).Where("version", trigger.Version).Updates(map[string]interface{}{
      "take_price": trigger.TakePrice,
      "stop_price": trigger.StopPrice,
      "timestamp":  trigger.Timestamp,
      "version":    gorm.Expr("version + ?", 1),
    }).Error
    if err != nil {
      return err
    }
  }

  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return err
  }

  if entryPrice > 0 {
    if trigger.Side == 1 && price > entryPrice {
      return errors.New(fmt.Sprintf("[%s] %s price big than entry price", trigger.Symbol, positionSide))
    }
    if trigger.Side == 2 && price < entryPrice {
      return errors.New(fmt.Sprintf("[%s] %s price small than  entry price", trigger.Symbol, positionSide))
    }
  }

  if trigger.Side == 1 && trigger.Price < price || trigger.Side == 2 && trigger.Price > price {
    var scalping *futuresModels.Scalping
    result = r.Db.Where("symbol=? AND side=?", trigger.Symbol, trigger.Side).Take(&scalping)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      var total int64
      r.Db.Model(&models.Scalping{}).Where("scalping_id=? AND status IN ?", scalping.ID, []int{1, 2}).Count(&total)
      if total < 1 {
        return errors.New(fmt.Sprintf("[%s] %s waiting for the scalping", trigger.Symbol, positionSide))
      }
    }
  }

  ipart, _ := math.Modf(trigger.Capital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }
  capital, err := r.PositionRepository.Capital(trigger.Capital, entryAmount, places)
  if err != nil {
    return errors.New("reach the max invest capital")
  }
  ratio := r.PositionRepository.Ratio(capital, entryAmount)
  buyAmount, _ := decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
  if buyAmount < notional {
    buyAmount = notional
  }

  var buyQuantity float64
  if entryAmount == 0 {
    buyAmount = notional
    buyQuantity, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(trigger.Price)).Float64()
  } else {
    buyQuantity = r.PositionRepository.BuyQuantity(trigger.Side, buyAmount, entryPrice, entryAmount)
  }

  buyPrice, _ := decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()

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
    return errors.New(fmt.Sprintf("[%s] %s price must reach %v", trigger.Symbol, positionSide, buyPrice))
  }

  if trigger.Side == 2 && price < buyPrice {
    return errors.New(fmt.Sprintf("[%s] %s price must reach %v", trigger.Symbol, positionSide, buyPrice))
  }

  buyQuantity, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyPrice)).Float64()
  buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
  entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()

  buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
  entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()

  if entryPrice == 0 {
    entryPrice = buyPrice
  } else {
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
  }

  sellPrice := r.PositionRepository.SellPrice(trigger.Side, entryPrice, entryAmount)
  if trigger.Side == 1 {
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if !r.CanBuy(trigger, buyPrice) {
    return errors.New(fmt.Sprintf("[%s] can not buy now", trigger.Symbol))
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return err
  }

  if balance["free"] < math.Max(balance["lock"], notional/float64(position.Leverage)) {
    return errors.New(fmt.Sprintf("[%s] free not enough", entity.QuoteAsset))
  }

  return r.Db.Transaction(func(tx *gorm.DB) (err error) {
    if position.ID != "" {
      result := tx.Model(&position).Where("version", position.Version).Updates(map[string]interface{}{
        "entry_quantity": gorm.Expr("entry_quantity + ?", buyQuantity),
        "version":        gorm.Expr("version + ?", 1),
      })
      if result.Error != nil {
        return result.Error
      }
      if result.RowsAffected == 0 {
        return errors.New("position update failed")
      }
    }

    orderID, err := r.OrdersRepository.Create(trigger.Symbol, positionSide, side, buyPrice, buyQuantity)
    if err != nil {
      _, ok := err.(common.APIError)
      if ok {
        return err
      }
      tx.Model(&trigger).Where("version", trigger.Version).Updates(map[string]interface{}{
        "remark":  err.Error(),
        "version": gorm.Expr("version + ?", 1),
      })
    }

    trading := models.Trigger{
      ID:           xid.New().String(),
      Symbol:       trigger.Symbol,
      TriggerID:    trigger.ID,
      BuyOrderId:   orderID,
      BuyPrice:     buyPrice,
      BuyQuantity:  buyQuantity,
      SellPrice:    sellPrice,
      SellQuantity: buyQuantity,
    }
    return tx.Create(&trading).Error
  })
}

func (r *TriggersRepository) Flush(id string) error {
  var trigger *futuresModels.Trigger
  result := r.Db.First(&trigger, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("trigger empty")
  }

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(trigger, price)
  if err != nil {
    log.Println("take error", trigger.Symbol, err)
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

  var tradings []*models.Trigger
  r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 2}).Find(&tradings)

  for _, trading := range tradings {
    if trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      timestamp := trading.CreatedAt.Unix()
      if trading.BuyOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, positionSide, side, trading.BuyQuantity, timestamp-30)
        if orderID > 0 {
          trading.BuyOrderId = orderID
          result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "buy_order_id": trading.BuyOrderId,
            "version":      gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("order update failed")
          }
        }
        if timestamp < time.Now().Unix()-900 {
          r.Db.Model(&trading).Update("status", 6)
        }
      } else {
        if timestamp < time.Now().Unix()-900 {
          if status == "NEW" {
            r.OrdersRepository.Cancel(trading.Symbol, trading.BuyOrderId)
          }
          if status == "" {
            r.OrdersRepository.Flush(trading.Symbol, trading.BuyOrderId)
          }
        }
      }

      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        continue
      }

      if status == "FILLED" {
        trading.Status = 1
      } else {
        trading.Status = 4
      }

      result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "buy_order_id": trading.BuyOrderId,
        "status":       trading.Status,
        "version":      gorm.Expr("version + ?", 1),
      })
      if result.Error != nil {
        return result.Error
      }
      if result.RowsAffected == 0 {
        return errors.New("order update failed")
      }
    }

    if trading.Status == 2 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.SellOrderId)
      timestamp := trading.UpdatedAt.Unix()
      if trading.SellOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, positionSide, side, trading.SellQuantity, timestamp-30)
        if orderID > 0 {
          trading.SellOrderId = orderID
          result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "sell_order_id": trading.SellOrderId,
            "version":       gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("order update failed")
          }
        }
        if timestamp < time.Now().Unix()-900 {
          r.Db.Model(&trading).Update("status", 1)
        }
      } else {
        if timestamp < time.Now().Unix()-900 {
          if status == "NEW" {
            r.OrdersRepository.Cancel(trading.Symbol, trading.SellOrderId)
          }
          if status == "" {
            r.OrdersRepository.Flush(trading.Symbol, trading.SellOrderId)
          }
        }
      }

      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        continue
      }

      if status == "FILLED" {
        trading.Status = 3
      } else if status == "CANCELED" {
        trading.SellOrderId = 0
        trading.Status = 1
      } else {
        trading.Status = 5
      }

      result := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "sell_order_id": trading.SellOrderId,
        "status":        trading.Status,
        "version":       gorm.Expr("version + ?", 1),
      })
      if result.Error != nil {
        return result.Error
      }
      if result.RowsAffected == 0 {
        return errors.New("order update failed")
      }
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

  position, err := r.PositionRepository.Get(trigger.Symbol, trigger.Side)
  if err != nil {
    return err
  }

  if position.EntryQuantity == 0 {
    r.Close(trigger)
    return errors.New(fmt.Sprintf("[%s] %s empty position", trigger.Symbol, positionSide))
  }

  if position.Timestamp > trigger.Timestamp {
    trigger.Timestamp = position.Timestamp
  }

  entity, err := r.SymbolsRepository.Get(trigger.Symbol)
  if err != nil {
    return err
  }

  tickSize, _, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  entryPrice := position.EntryPrice

  var sellPrice float64
  var trading *models.Trigger

  if trigger.Side == 1 {
    result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Order("sell_price asc").Take(&trading)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New(fmt.Sprintf("[%s] %s empty trading", trigger.Symbol, positionSide))
    }
    if price < trading.SellPrice {
      if price < entryPrice*1.0138 {
        return errors.New("price too low")
      }
      sellPrice = entryPrice * 1.0138
    } else {
      sellPrice = trading.SellPrice
    }
    if sellPrice < price*0.9985 {
      sellPrice = price * 0.9985
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if trigger.Side == 2 {
    result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Order("sell_price desc").Take(&trading)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("empty trading")
    }
    if price > trading.SellPrice {
      if price > entryPrice*0.9862 {
        return errors.New("price too high")
      }
      sellPrice = entryPrice * 0.9862
    } else {
      if price > entryPrice*0.9862 {
        return errors.New("price too high")
      }
      if price > entryPrice*0.9862 {
        return errors.New("price too high")
      }
      sellPrice = trading.SellPrice
    }
    if sellPrice > price*1.0015 {
      sellPrice = price * 1.0015
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  orderID, err := r.OrdersRepository.Create(trading.Symbol, positionSide, side, sellPrice, trading.SellQuantity)
  if err != nil {
    _, ok := err.(common.APIError)
    if ok {
      return err
    }
    r.Db.Model(&trigger).Where("version", trigger.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
    "sell_order_id": orderID,
    "status":        2,
    "version":       gorm.Expr("version + ?", 1),
  })

  return nil
}

func (r *TriggersRepository) Close(trigger *futuresModels.Trigger) {
  var total int64
  r.Db.Model(&models.Trigger{}).Where("trigger_id = ? AND status IN ?", trigger.ID, []int{0, 1, 2}).Count(&total)
  if total == 0 {
    return
  }
  r.Db.Model(&models.Trigger{}).Where("trigger_id = ? AND status IN ?", trigger.ID, []int{0, 2}).Count(&total)
  if total > 0 {
    return
  }
  timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
  if trigger.Timestamp > timestamp {
    return
  }
  r.Db.Model(&models.Trigger{}).Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Update("status", 5)
}

func (r *TriggersRepository) CanBuy(
  trigger *futuresModels.Trigger,
  price float64,
) bool {
  var trading models.Trigger
  if trigger.Side == 1 {
    result := r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Order("buy_price asc").Take(&trading)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if trading.Status == 0 {
        return false
      }
      if price >= trading.BuyPrice {
        return false
      }
    }
  }
  if trigger.Side == 2 {
    result := r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Order("buy_price desc").Take(&trading)
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if trading.Status == 0 {
        return false
      }
      if price <= trading.BuyPrice {
        return false
      }
    }
  }
  return true
}
