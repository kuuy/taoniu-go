package tradings

import (
  "context"
  "errors"
  "fmt"
  "log"
  "math"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot"
  tradingsModels "taoniu.local/cryptos/models/binance/spot/tradings"
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
  r.Db.Model(&models.Trigger{}).Where("status", 1).Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Ids() []string {
  var ids []string
  r.Db.Model(&models.Trigger{}).Where("status", 1).Pluck("id", &ids)
  return ids
}

func (r *TriggersRepository) TriggerIds() []string {
  var ids []string
  r.Db.Model(&tradingsModels.Trigger{}).Select("trigger_id").Where("status", []int{0, 1, 2}).Distinct("trigger_id").Find(&ids)
  return ids
}

func (r *TriggersRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&tradingsModels.Trigger{})
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

func (r *TriggersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*tradingsModels.Trigger {
  var tradings []*tradingsModels.Trigger
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
  var trigger *models.Trigger
  result := r.Db.Take(&trigger, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("trigger not found")
    return
  }

  if trigger.ExpiredAt.Unix() < time.Now().Unix() {
    r.Db.Model(&trigger).Update("status", 4)
    return errors.New("trigger expired")
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

  var capital float64
  var quantity float64
  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64
  var sellPrice float64

  var cachedEntryPrice float64
  var cachedEntryQuantity float64

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_TRIGGERS_PLACE, trigger.Symbol)
  values, _ := r.Rdb.HMGet(r.Ctx, redisKey, []string{
    "entry_price",
    "entry_quantity",
    "buy_price",
    "buy_quantity",
  }...).Result()
  if len(values) == 4 && values[0] != nil && values[1] != nil {
    cachedEntryPrice, _ = strconv.ParseFloat(values[0].(string), 64)
    cachedEntryQuantity, _ = strconv.ParseFloat(values[1].(string), 64)
  }

  if cachedEntryPrice == entryPrice && cachedEntryQuantity == entryQuantity {
    buyPrice, _ = strconv.ParseFloat(values[2].(string), 64)
    buyQuantity, _ = strconv.ParseFloat(values[3].(string), 64)
    buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
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
        err = errors.New(fmt.Sprintf("trigger [%s] reach the max invest capital", trigger.Symbol))
        return
      }
      ratio := r.PositionRepository.Ratio(capital, entryAmount)
      buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
      if buyAmount < notional {
        buyAmount = notional
      }
      buyQuantity = r.PositionRepository.BuyQuantity(buyAmount, entryPrice, entryAmount)
      buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
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
    err = errors.New(fmt.Sprintf("trigger [%s] price %v is negative", trigger.Symbol, buyPrice))
    return
  }

  if buyQuantity < 0 {
    err = errors.New(fmt.Sprintf("trigger [%s] quantity %v is negative", trigger.Symbol, buyQuantity))
    return
  }

  if price < buyPrice {
    buyPrice = price
  }
  buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
  entryQuantity, _ = decimal.NewFromFloat(position.EntryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
  entryAmount, _ = decimal.NewFromFloat(position.EntryPrice).Mul(decimal.NewFromFloat(position.EntryQuantity)).Add(decimal.NewFromFloat(buyAmount)).Float64()
  entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
  sellPrice = r.PositionRepository.SellPrice(entryPrice, entryAmount)
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
    err = errors.New(fmt.Sprintf("triggers free balance must reach %v", math.Max(buyAmount, config.TRIGGERS_MIN_BINANCE)))
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

  orderId, err := r.OrdersRepository.Create(trigger.Symbol, side, buyPrice, buyQuantity)
  if err != nil {
    if common.IsBinanceAPIError(err) {
      return
    }
    r.Db.Model(&trigger).Where("version", trigger.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  trading := tradingsModels.Trigger{
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
  var trigger *models.Trigger
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

  placeSide := "BUY"
  takeSide := "SELL"

  var tradings []*tradingsModels.Trigger
  r.Db.Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 2}).Find(&tradings)

  timestamp := time.Now().Add(-15 * time.Minute).Unix()

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, trigger.Symbol)

  for _, trading := range tradings {
    if trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      if trading.BuyOrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, placeSide, trading.BuyQuantity, trading.UpdatedAt.Add(-120*time.Second).Unix())
        if orderId > 0 {
          status = r.OrdersRepository.Status(trading.Symbol, orderId)
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
        } else {
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
            r.Rdb.Del(r.Ctx, redisKey)
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
        r.Rdb.Set(r.Ctx, redisKey, trading.BuyPrice, time.Hour*24)
      } else if status == "CANCELED" {
        result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
          "status":  4,
          "version": gorm.Expr("version + ?", 1),
        })
        if result.Error != nil {
          return result.Error
        }
        if result.RowsAffected == 0 {
          return errors.New("order update failed")
        }
        r.Rdb.Del(r.Ctx, redisKey)
      }
    }

    if trading.Status == 2 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.SellOrderId)
      if trading.SellOrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, takeSide, trading.SellQuantity, trading.UpdatedAt.Add(-120*time.Second).Unix())
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
        result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
          "status":  3,
          "version": gorm.Expr("version + ?", 1),
        })
        if result.Error != nil {
          return result.Error
        }
        if result.RowsAffected == 0 {
          return errors.New("order update failed")
        }
        r.Rdb.Del(r.Ctx, redisKey)
      } else if status == "CANCELED" {
        result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
          "sell_order_id": 0,
          "status":        1,
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
  }

  return
}

func (r *TriggersRepository) Take(trigger *models.Trigger, price float64) (err error) {
  var side = "SELL"
  var entryPrice float64
  var sellPrice float64
  var trading *tradingsModels.Trigger

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
      r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, trigger.Symbol))
    }
    err = errors.New(fmt.Sprintf("[%s] empty position", trigger.Symbol))
    return
  }

  entryPrice = position.EntryPrice
  if trigger.Timestamp < position.Timestamp {
    trigger.Timestamp = position.Timestamp
  }

  entity, err := r.SymbolsRepository.Get(trigger.Symbol)
  if err != nil {
    return
  }

  tickSize, _, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  result := r.Db.Where("trigger_id=? AND status=?", trigger.ID, 1).Take(&trading)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty trigger to take")
  }
  if price < trading.SellPrice {
    if price < entryPrice*1.0105 {
      return errors.New("compare with entry price too low")
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

  balance, err := r.AccountRepository.Balance(entity.BaseAsset)
  if err != nil {
    return
  }

  if balance["free"] < trading.SellQuantity {
    err = errors.New(fmt.Sprintf("[%s] free not enough", entity.BaseAsset))
    return
  }

  orderId, err := r.OrdersRepository.Create(trading.Symbol, side, sellPrice, trading.SellQuantity)
  if err != nil {
    if common.IsBinanceAPIError(err) {
      return
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

  return
}

func (r *TriggersRepository) Close(trigger *models.Trigger) {
  var total int64
  r.Db.Model(&tradingsModels.Trigger{}).Where("trigger_id = ? AND status IN ?", trigger.ID, []int{0, 1}).Count(&total)
  if total == 0 {
    return
  }

  var tradings []*tradingsModels.Trigger
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
  trigger *models.Trigger,
  price float64,
) bool {
  var buyPrice float64
  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, trigger.Symbol)
  val, _ := r.Rdb.Get(r.Ctx, redisKey).Result()
  if val != "" {
    buyPrice, _ = strconv.ParseFloat(val, 64)
    if price >= buyPrice*0.9615 {
      return false
    }
  }

  isChange := false

  var tradings []*tradingsModels.Trigger
  r.Db.Select([]string{"status", "buy_price"}).Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Find(&tradings)
  for _, trading := range tradings {
    if trading.Status == 0 {
      return false
    }
    if price >= trading.BuyPrice*0.9615 {
      return false
    }
    if buyPrice == 0 || buyPrice > trading.BuyPrice {
      buyPrice = trading.BuyPrice
      isChange = true
    }
  }

  if isChange {
    r.Rdb.Set(r.Ctx, redisKey, buyPrice, time.Hour*24)
  }

  return true
}
