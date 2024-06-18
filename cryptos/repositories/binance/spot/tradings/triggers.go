package tradings

import (
  "context"
  "errors"
  "fmt"
  "log"
  "math"
  "time"

  apiCommon "github.com/adshao/go-binance/v2/common"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/tradings"
)

type TriggersRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  SymbolsRepository  SymbolsRepository
  AccountRepository  AccountRepository
  OrdersRepository   OrdersRepository
  PositionRepository PositionRepository
}

func (r *TriggersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&spotModels.Trigger{}).Where("status", 1).Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Ids() []string {
  var ids []string
  r.Db.Model(&spotModels.Trigger{}).Where("status", 1).Pluck("id", &ids)
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

func (r *TriggersRepository) Place(id string) (err error) {
  var trigger *spotModels.Trigger
  result := r.Db.Take(&trigger, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("trigger not found")
    return
  }

  var side = "BUY"

  position, err := r.PositionRepository.Get(trigger.Symbol)
  if err != nil {
    return
  }

  if position.EntryQuantity == 0 {
    err = errors.New(fmt.Sprintf("trigger [%s] empty position", trigger.Symbol))
    return
  }

  entity, err := r.SymbolsRepository.Get(trigger.Symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, notional, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  entryPrice := position.EntryPrice
  entryQuantity := position.EntryQuantity
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return
  }

  if price > entryPrice {
    err = errors.New(fmt.Sprintf("trigger [%s] price big than entry price", trigger.Symbol))
    return
  }

  ipart, _ := math.Modf(trigger.Capital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }
  capital, err := r.PositionRepository.Capital(trigger.Capital, entryAmount, places)
  if err != nil {
    err = errors.New("reach the max invest capital")
    return
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
    buyQuantity = r.PositionRepository.BuyQuantity(buyAmount, entryPrice, entryAmount)
  }

  buyPrice, _ := decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()

  if price < buyPrice {
    buyPrice = price
  }

  buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
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

  sellPrice := r.PositionRepository.SellPrice(entryPrice, entryAmount)
  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

  if price > buyPrice {
    err = errors.New(fmt.Sprintf("trigger [%s] price must reach %v", trigger.Symbol, buyPrice))
    return
  }

  if !r.CanBuy(trigger, buyPrice) {
    err = errors.New(fmt.Sprintf("trigger [%s] can not buy now", trigger.Symbol))
    return
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return
  }

  if balance["free"] < math.Max(buyAmount, config.TRIGGERS_MIN_BINANCE) {
    err = errors.New(fmt.Sprintf("trigger [%s] free not enough", entity.Symbol))
    return
  }

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_PLACE, trigger.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  orderID, err := r.OrdersRepository.Create(trigger.Symbol, side, buyPrice, buyQuantity)
  if err != nil {
    _, ok := err.(apiCommon.APIError)
    if ok {
      return
    }
    r.Db.Model(&trigger).Where("version", trigger.Version).Updates(map[string]interface{}{
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
    Version:      1,
  }
  r.Db.Create(&trading)

  return
}

func (r *TriggersRepository) Flush(id string) (err error) {
  var trigger *spotModels.Trigger
  result := r.Db.Take(&trigger, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("trigger not found")
  }

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return
  }
  err = r.Take(trigger, price)
  if err != nil {
    log.Println("take error", trigger.Symbol, err)
  }

  var side = "BUY"

  var tradings []*models.Trigger
  r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 2}).Find(&tradings)

  timestamp := time.Now().Add(-15 * time.Minute).Unix()

  for _, trading := range tradings {
    if trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      if trading.BuyOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, side, trading.BuyQuantity, trading.UpdatedAt.Add(-120*time.Second).Unix())
        if orderID > 0 {
          status = r.OrdersRepository.Status(trading.Symbol, orderID)
          trading.BuyOrderId = orderID
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
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
        if trading.UpdatedAt.Unix() < timestamp {
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "status":  6,
            "version": gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("order update failed")
          }
        }
      } else {
        if trading.BuyOrderId > 0 && trading.UpdatedAt.Unix() < timestamp {
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
      } else if status == "CANCELED" {
        trading.Status = 4
      }

      result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
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
      if trading.SellOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, side, trading.SellQuantity, trading.UpdatedAt.Add(-120*time.Second).Unix())
        if orderID > 0 {
          trading.SellOrderId = orderID
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "sell_order_id": trading.SellOrderId,
            "version":       gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("order update failed")
          }
        } else {
          if trading.UpdatedAt.Unix() < timestamp {
            result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
              "status":  1,
              "version": gorm.Expr("version + ?", 1),
            })
            if result.Error != nil {
              return result.Error
            }
            if result.RowsAffected == 0 {
              return errors.New("order update failed")
            }
          }
        }
      } else {
        if trading.UpdatedAt.Unix() < timestamp {
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
      }

      result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
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

  return
}

func (r *TriggersRepository) Take(trigger *spotModels.Trigger, price float64) error {
  var side = "SELL"
  var entryPrice float64
  var sellPrice float64
  var trading *models.Trigger

  position, err := r.PositionRepository.Get(trigger.Symbol)
  if err != nil {
    return err
  }

  if position.EntryQuantity == 0 {
    timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    if position.Timestamp > timestamp {
      return errors.New("waiting for more time")
    }
    if position.Timestamp > trigger.Timestamp+9e8 {
      r.Close(trigger)
    }
    return errors.New(fmt.Sprintf("[%s] empty position", trigger.Symbol))
  }

  entryPrice = position.EntryPrice
  if trigger.Timestamp < position.Timestamp {
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

  result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Order("sell_price asc").Take(&trading)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New(fmt.Sprintf("[%s] empty trading", trigger.Symbol))
  }
  if price < trading.SellPrice {
    if price < entryPrice*1.0105 {
      return errors.New("compare with sell price too low")
    }
    timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    if trading.UpdatedAt.UnixMicro() > timestamp {
      return errors.New("waiting for more time")
    }
    sellPrice = entryPrice * 1.0105
  } else {
    if entryPrice > trading.SellPrice {
      if price < entryPrice*1.0105 {
        return errors.New("compare with entry price too low")
      }
      sellPrice = entryPrice * 1.0105
    } else {
      sellPrice = trading.SellPrice
    }
  }
  if sellPrice < price*0.9985 {
    sellPrice = price * 0.9985
  }
  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()

  orderID, err := r.OrdersRepository.Create(trading.Symbol, side, sellPrice, trading.SellQuantity)
  if err != nil {
    _, ok := err.(apiCommon.APIError)
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

func (r *TriggersRepository) Close(trigger *spotModels.Trigger) {
  var total int64
  var tradings []*models.Trigger
  r.Db.Model(&tradings).Where("trigger_id = ? AND status IN ?", trigger.ID, []int{0, 1}).Count(&total)
  if total == 0 {
    return
  }
  r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Find(&tradings)
  timestamp := time.Now().Add(-30 * time.Minute).UnixMicro()
  for _, trading := range tradings {
    if trading.UpdatedAt.UnixMicro() < timestamp {
      r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "status":  5,
        "version": gorm.Expr("version + ?", 1),
      })
    }
  }
}

func (r *TriggersRepository) CanBuy(
  trigger *spotModels.Trigger,
  price float64,
) bool {
  var trading models.Trigger
  result := r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Order("buy_price asc").Take(&trading)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if trading.Status == 0 {
      return false
    }
    if price >= trading.BuyPrice*0.9615 {
      return false
    }
  }
  return true
}
