package tradings

import (
  "context"
  "errors"
  "fmt"
  "log"
  "math"
  "strconv"
  "time"

  apiCommon "github.com/adshao/go-binance/v2/common"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/margin/cross"
  crossModels "taoniu.local/cryptos/models/binance/margin/cross"
  models "taoniu.local/cryptos/models/binance/margin/cross/tradings"
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
  r.Db.Model(&crossModels.Trigger{}).Where("status", 1).Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Ids() []string {
  var ids []string
  r.Db.Model(&crossModels.Trigger{}).Where("status", 1).Pluck("id", &ids)
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
  var trigger *crossModels.Trigger
  result := r.Db.First(&trigger, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("trigger not found")
    return
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
    return
  }

  if position.EntryQuantity == 0 {
    return errors.New(fmt.Sprintf("trigger [%s] %s empty position", trigger.Symbol, positionSide))
  }

  entity, err := r.SymbolsRepository.Get(trigger.Symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, notional, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  entryPrice := position.EntryPrice
  entryQuantity := position.EntryQuantity
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return
  }

  if trigger.Side == 1 && price > entryPrice {
    err = errors.New(fmt.Sprintf("trigger [%s] %s price big than entry price", trigger.Symbol, positionSide))
    return
  }
  if trigger.Side == 2 && price < entryPrice {
    err = errors.New(fmt.Sprintf("trigger [%s] %s price small than  entry price", trigger.Symbol, positionSide))
    return
  }

  var capital float64
  var quantity float64
  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64
  var sellPrice float64

  var cachedEntryPrice float64
  var cachedEntryQuantity float64

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_TRIGGERS_PLACE, positionSide, trigger.Symbol)
  values, _ := r.Rdb.HMGet(r.Ctx, redisKey, []string{
    "entry_price",
    "entry_quantity",
    "buy_price",
    "buy_quantity",
  }...).Result()
  if values[0] != nil {
    cachedEntryPrice, _ = strconv.ParseFloat(values[0].(string), 64)
    cachedEntryQuantity, _ = strconv.ParseFloat(values[1].(string), 64)
  }

  if cachedEntryPrice == entryPrice && cachedEntryQuantity == entryQuantity {
    buyPrice, _ = strconv.ParseFloat(values[2].(string), 64)
    buyQuantity, _ = strconv.ParseFloat(values[3].(string), 64)
    log.Println("load from cached price", buyPrice, buyQuantity)
  } else {
    cachedEntryPrice = entryPrice
    cachedEntryQuantity = entryQuantity

    ipart, _ := math.Modf(trigger.Capital)
    places := 1
    for ; ipart >= 10; ipart = ipart / 10 {
      places++
    }

    for i := 0; i < 2; i++ {
      capital, err = r.PositionRepository.Capital(trigger.Capital, entryAmount, places)
      if err != nil {
        return errors.New("reach the max invest capital")
      }
      ratio := r.PositionRepository.Ratio(capital, entryAmount)
      buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
      if buyAmount < notional {
        buyAmount = notional
      }
      buyQuantity = r.PositionRepository.BuyQuantity(trigger.Side, buyAmount, entryPrice, entryAmount)
      buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
      if trigger.Side == 1 {
        buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
      } else {
        buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
      }
      buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
      buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
      entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
      entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
      entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
      if i == 0 {
        quantity = buyQuantity
      }
    }
    buyQuantity = quantity

    r.Rdb.HMSet(
      r.Ctx,
      redisKey,
      map[string]interface{}{
        "entry_price":    cachedEntryPrice,
        "entry_quantity": cachedEntryQuantity,
        "buy_price":      buyPrice,
        "buy_quantity":   buyQuantity,
      },
    )
  }

  if buyPrice < 0 {
    err = errors.New(fmt.Sprintf("trigger [%s] %s price %v is negative", trigger.Symbol, positionSide, buyPrice))
    return
  }

  if buyQuantity < 0 {
    err = errors.New(fmt.Sprintf("trigger [%s] %s quantity %v is negative", trigger.Symbol, positionSide, buyQuantity))
    return
  }

  if trigger.Side == 1 && price < buyPrice {
    buyPrice = price
  } else if trigger.Side == 2 && price > buyPrice {
    buyPrice = price
  }
  buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
  entryQuantity, _ = decimal.NewFromFloat(position.EntryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
  entryAmount, _ = decimal.NewFromFloat(position.EntryPrice).Mul(decimal.NewFromFloat(position.EntryQuantity)).Add(decimal.NewFromFloat(buyAmount)).Float64()
  entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
  sellPrice = r.PositionRepository.SellPrice(trigger.Side, entryPrice, entryAmount)
  if trigger.Side == 1 {
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if trigger.Side == 1 && price > buyPrice {
    err = errors.New(fmt.Sprintf("trigger [%s] %s price must reach %v", trigger.Symbol, positionSide, buyPrice))
    return
  }

  if trigger.Side == 2 && price < buyPrice {
    err = errors.New(fmt.Sprintf("trigger [%s] %s price must reach %v", trigger.Symbol, positionSide, buyPrice))
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

  if balance["borrowed"]+buyAmount > config.TRIGGERS_MAX_BORROWED {
    err = errors.New(fmt.Sprintf("[%s] over borrowed", entity.Symbol))
    return
  }

  if balance["free"] < buyAmount {
    mutex := common.NewMutex(
      r.Rdb,
      r.Ctx,
      fmt.Sprintf(config.LOCKS_ACCOUNT_BORROW, trigger.Symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      return nil
    }
    var transferId int64
    transferId, err = r.AccountRepository.Borrow(entity.QuoteAsset, buyAmount)
    if err != nil {
      if _, ok := err.(apiCommon.APIError); ok {
        mutex.Unlock()
        return
      }
      log.Println("error", err)
    }
    mutex.Unlock()
    log.Println("loan", transferId, buyAmount)
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

  orderId, err := r.OrdersRepository.Create(trigger.Symbol, side, buyPrice, buyQuantity)
  if err != nil {
    if _, ok := err.(apiCommon.APIError); ok {
      return
    }
    r.Db.Model(&trigger).Where("version", trigger.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  r.Rdb.Set(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, positionSide, trigger.Symbol), buyPrice, -1)

  trading := models.Trigger{
    ID:           xid.New().String(),
    Symbol:       trigger.Symbol,
    TriggerID:    trigger.ID,
    BuyOrderId:   orderId,
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
  var trigger *crossModels.Trigger
  result := r.Db.First(&trigger, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("trigger not found")
    return
  }

  price, err := r.SymbolsRepository.Price(trigger.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(trigger, price)
  if err != nil {
    log.Println("take error", trigger.Symbol, err)
  }

  var placeSide string
  var takeSide string
  if trigger.Side == 1 {
    placeSide = "BUY"
    takeSide = "SELL"
  } else if trigger.Side == 2 {
    placeSide = "SELL"
    takeSide = "SELL"
  }

  var tradings []*models.Trigger
  r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 2}).Find(&tradings)

  for _, trading := range tradings {
    if trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      timestamp := trading.CreatedAt.Unix()
      if trading.BuyOrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, placeSide, trading.BuyQuantity, timestamp-30)
        if orderId > 0 {
          trading.BuyOrderId = orderId
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
      timestamp := trading.UpdatedAt.Unix()
      if trading.SellOrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, takeSide, trading.SellQuantity, timestamp-30)
        if orderId > 0 {
          trading.SellOrderId = orderId
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

func (r *TriggersRepository) Take(trigger *crossModels.Trigger, price float64) error {
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
    timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    if position.Timestamp > timestamp {
      return errors.New("waiting for more time")
    }
    if position.Timestamp > trigger.Timestamp {
      r.Close(trigger)
    }
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
    result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Take(&trading)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New(fmt.Sprintf("[%s] %s empty trading", trigger.Symbol, positionSide))
    }
    if price < trading.SellPrice {
      if price < entryPrice*1.0105 {
        return errors.New("price too low")
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
  }

  if trigger.Side == 2 {
    result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Order("sell_price desc").Take(&trading)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("empty trading")
    }
    if price > trading.SellPrice {
      if price > entryPrice*0.9895 {
        return errors.New("compare with entry price too high")
      }
      timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
      if trading.UpdatedAt.UnixMicro() > timestamp {
        return errors.New("waiting for more time")
      }
      sellPrice = entryPrice * 0.9895
    } else {
      if entryPrice < trading.SellPrice {
        if price > entryPrice*0.9895 {
          return errors.New("compare with entry price too high")
        }
        sellPrice = entryPrice * 0.9895
      } else {
        sellPrice = trading.SellPrice
      }
    }
    if sellPrice > price*1.0015 {
      sellPrice = price * 1.0015
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  orderId, err := r.OrdersRepository.Create(trading.Symbol, side, sellPrice, trading.SellQuantity)
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
    "sell_order_id": orderId,
    "status":        2,
    "version":       gorm.Expr("version + ?", 1),
  })

  return nil
}

func (r *TriggersRepository) Close(trigger *crossModels.Trigger) {
  var total int64
  r.Db.Model(&models.Trigger{}).Where("trigger_id = ? AND status IN ?", trigger.ID, []int{0, 1}).Count(&total)
  if total == 0 {
    return
  }

  var tradings []*models.Trigger
  r.Db.Select([]string{"id", "version", "updated_at"}).Where("trigger_id=? AND status=?", trigger.ID, 1).Find(&tradings)
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
  trigger *crossModels.Trigger,
  price float64,
) bool {
  var tradings []*models.Trigger
  r.Db.Select([]string{"status", "buy_price"}).Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Find(&tradings)
  for _, trading := range tradings {
    if trading.Status == 0 {
      return false
    }
    if trigger.Side == 1 && price >= trading.BuyPrice*0.9615 {
      return false
    }
    if trigger.Side == 2 && price <= trading.BuyPrice*1.0385 {
      return false
    }
  }
  return true
}
