package gambling

import (
  "context"
  "errors"
  "fmt"
  "gorm.io/datatypes"
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
  config "taoniu.local/cryptos/config/binance/futures"
  gamblingModels "taoniu.local/cryptos/models/binance/futures/gambling"
  tradingsModels "taoniu.local/cryptos/models/binance/futures/tradings/gambling"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type AntRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  SymbolsRepository  SymbolsRepository
  AccountRepository  AccountRepository
  OrdersRepository   OrdersRepository
  PositionRepository PositionRepository
  GamblingRepository *repositories.GamblingRepository
}

func (r *AntRepository) Ids() []string {
  var ids []string
  r.Db.Model(&gamblingModels.Ant{}).Where("status=?", 1).Pluck("id", &ids)
  return ids
}

func (r *AntRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&tradingsModels.Ant{})
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

func (r *AntRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*tradingsModels.Ant {
  var tradings []*tradingsModels.Ant
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "ant_id",
    "mode",
    "price",
    "quantity",
    "order_id",
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

func (r *AntRepository) Flush(id string) (err error) {
  var ant *gamblingModels.Ant
  var result = r.Db.First(&ant, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty gambling ant to flush")
  }

  price, err := r.SymbolsRepository.Price(ant.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(ant, price)
  if err != nil {
    log.Println("take error", ant.Symbol, err)
  }

  var positionSide string
  var placeSide string
  var takeSide string
  if ant.Side == 1 {
    positionSide = "LONG"
    placeSide = "BUY"
    takeSide = "SELL"
  } else if ant.Side == 2 {
    positionSide = "SHORT"
    placeSide = "SELL"
    takeSide = "BUY"
  }

  var tradings []*tradingsModels.Ant
  r.Db.Where("ant_id=? AND status IN ?", ant.ID, []int{0, 1}).Find(&tradings)

  timestamp := time.Now().Add(-15 * time.Minute).Unix()

  for _, trading := range tradings {
    if trading.Mode == 1 && trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.OrderId)
      if trading.OrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, positionSide, placeSide, trading.Quantity, trading.UpdatedAt.Add(-120*time.Second).UnixMilli())
        if orderId > 0 {
          status = r.OrdersRepository.Status(trading.Symbol, orderId)
          trading.OrderId = orderId
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "order_id": trading.OrderId,
            "version":  gorm.Expr("version + ?", 1),
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
            r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode))
          }
        }
      } else {
        if trading.OrderId > 0 && trading.UpdatedAt.Unix() < timestamp {
          if status == "NEW" {
            r.OrdersRepository.Cancel(trading.Symbol, trading.OrderId)
          }
          if status == "" {
            log.Println("order flush", trading.Symbol, trading.OrderId)
            r.OrdersRepository.Flush(trading.Symbol, trading.OrderId)
          }
        }
      }

      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        continue
      }

      if status == "FILLED" {
        err = r.Db.Transaction(func(tx *gorm.DB) (err error) {
          placeQuantity, _ := decimal.NewFromFloat(ant.PlaceQuantity).Add(decimal.NewFromFloat(trading.Quantity)).Float64()
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "status":  1,
            "version": gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("trading update failed")
          }
          result = r.Db.Model(&ant).Where("version", ant.Version).Updates(map[string]interface{}{
            "place_quantity": placeQuantity,
            "version":        gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("gambling ant update failed")
          }
          return
        })
        if err != nil {
          return
        }
        r.Rdb.Set(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode), trading.Price, -1)
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
        r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode))
      }
    }

    if trading.Mode == 2 && trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.OrderId)
      if trading.OrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, positionSide, takeSide, trading.Quantity, trading.UpdatedAt.Add(-120*time.Second).UnixMilli())
        if orderId > 0 {
          status = r.OrdersRepository.Status(trading.Symbol, orderId)
          trading.OrderId = orderId
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "order_id": trading.OrderId,
            "version":  gorm.Expr("version + ?", 1),
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
        if trading.OrderId > 0 && trading.UpdatedAt.Unix() < timestamp {
          if status == "NEW" {
            r.OrdersRepository.Cancel(trading.Symbol, trading.OrderId)
          }
          if status == "" {
            r.OrdersRepository.Flush(trading.Symbol, trading.OrderId)
          }
        }
      }

      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        continue
      }

      if status == "FILLED" {
        err = r.Db.Transaction(func(tx *gorm.DB) (err error) {
          takeQuantity, _ := decimal.NewFromFloat(ant.TakeQuantity).Add(decimal.NewFromFloat(trading.Quantity)).Float64()
          result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "status":  1,
            "version": gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("trading update failed")
          }
          result = r.Db.Model(&ant).Where("version", ant.Version).Updates(map[string]interface{}{
            "take_prices":     datatypes.NewJSONSlice(ant.TakePrices[1:]),
            "take_quantities": datatypes.NewJSONSlice(ant.TakeQuantities[1:]),
            "take_quantity":   takeQuantity,
            "version":         gorm.Expr("version + ?", 1),
          })
          if result.Error != nil {
            return result.Error
          }
          if result.RowsAffected == 0 {
            return errors.New("ant update failed")
          }
          return
        })
        if err != nil {
          return
        }
        r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode))
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
        r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode))
      }
    }
  }

  return
}

func (r *AntRepository) Place(id string) (err error) {
  var ant *gamblingModels.Ant
  result := r.Db.First(&ant, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = errors.New("gambling ant not found")
    return
  }

  if ant.ExpiredAt.Unix() < time.Now().Unix() {
    r.Db.Model(&ant).Update("status", 4)
    return errors.New("gambling ant expired")
  }

  if ant.Mode != 1 {
    return errors.New("gambling ant has been placed")
  }

  if ant.PlaceQuantity == ant.EntryQuantity {
    r.Db.Model(&ant).Update("mode", 2)
    return errors.New("gambling ant has been placed")
  }

  var positionSide string
  var side string
  if ant.Side == 1 {
    positionSide = "LONG"
    side = "BUY"
  } else if ant.Side == 2 {
    positionSide = "SHORT"
    side = "SELL"
  }
  log.Println("side", side)

  position, err := r.PositionRepository.Get(ant.Symbol, ant.Side)
  if err != nil {
    return
  }

  if position.EntryQuantity == 0 {
    return errors.New(fmt.Sprintf("gambling ant place [%s] %s empty position", ant.Symbol, positionSide))
  }

  entity, err := r.SymbolsRepository.Get(ant.Symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  entryPrice := position.EntryPrice
  entryQuantity := position.EntryQuantity

  price, err := r.SymbolsRepository.Price(ant.Symbol)
  if err != nil {
    return
  }

  if ant.Side == 1 && price > entryPrice {
    err = errors.New(fmt.Sprintf("gambling ant place [%s] %s price big than entry price", ant.Symbol, positionSide))
    return
  }
  if ant.Side == 2 && price < entryPrice {
    err = errors.New(fmt.Sprintf("gambling ant place [%s] %s price small than entry price", ant.Symbol, positionSide))
    return
  }

  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64

  var cachedEntryPrice float64
  var cachedEntryQuantity float64

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_PLACE, positionSide, ant.Symbol)
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
    if ant.Side == 1 {
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    log.Println("load from cached price", buyPrice, buyQuantity)
  } else {
    cachedEntryPrice = entryPrice
    cachedEntryQuantity = entryQuantity

    var tradings []*tradingsModels.Ant
    r.Db.Select("price").Where("ant_id=? AND mode=1 AND status IN ?", ant.ID, []int{0, 1}).Find(&tradings)

    lastPrice := ant.PlacePrices[len(ant.PlacePrices)-1]
    for _, trading := range tradings {
      if buyPrice == 0.0 || ant.Side == 1 && buyPrice > trading.Price || ant.Side == 2 && buyPrice > trading.Price {
        buyPrice = trading.Price
      }
      if trading.Status == 1 {
        if ant.Side == 1 && trading.Price <= lastPrice || ant.Side == 2 && trading.Price >= lastPrice {
          r.Db.Model(&ant).Update("mode", 2)
          return errors.New("gambling ant has been placed")
        }
      }
    }

    for i, planPrice := range ant.PlacePrices {
      if buyPrice == 0.0 || ant.Side == 1 && buyPrice > planPrice || ant.Side == 2 && buyPrice > planPrice {
        buyPrice = planPrice
        buyQuantity = ant.PlaceQuantities[i]
        break
      }
    }

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
    err = errors.New(fmt.Sprintf("gambling ant [%s] %s price %v is negative", ant.Symbol, positionSide, buyPrice))
    return
  }

  if buyQuantity < 0 {
    err = errors.New(fmt.Sprintf("gambling ant [%s] %s quantity %v is negative", ant.Symbol, positionSide, buyQuantity))
    return
  }

  if ant.Side == 1 && price < buyPrice {
    buyPrice = price
  } else if ant.Side == 2 && price > buyPrice {
    buyPrice = price
  }

  if ant.Side == 1 && price > buyPrice {
    err = errors.New(fmt.Sprintf("gambling ant [%s] %s price must reach %v", ant.Symbol, positionSide, buyPrice))
    return
  }

  if ant.Side == 2 && price < buyPrice {
    err = errors.New(fmt.Sprintf("gambling ant [%s] %s price must reach %v", ant.Symbol, positionSide, buyPrice))
    return
  }

  buyAmount, _ = decimal.NewFromFloat(buyQuantity).Mul(decimal.NewFromFloat(buyPrice)).Float64()
  if buyAmount > config.GAMBLING_ANT_MAX_AMOUNT {
    return errors.New(fmt.Sprintf("gambling ant [%s] %s amount can not exceed %v", ant.Symbol, positionSide, config.GAMBLING_ANT_MAX_AMOUNT))
  }

  if !r.CanBuy(ant, buyPrice) {
    err = errors.New(fmt.Sprintf("gambling ant [%s] can not place now", ant.Symbol))
    return
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return
  }

  if balance["free"] < math.Max(buyAmount, config.GAMBLING_ANT_MIN_BINANCE) {
    err = errors.New(fmt.Sprintf("[%s] free not enough", entity.Symbol))
    return
  }

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_PLACE, ant.Symbol, ant.Side),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(ant.Symbol, positionSide, side, buyPrice, buyQuantity)
  if err != nil {
    _, ok := err.(apiCommon.APIError)
    if ok {
      return
    }
    r.Db.Model(&ant).Where("version", ant.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  r.Db.Create(&tradingsModels.Ant{
    ID:       xid.New().String(),
    Symbol:   ant.Symbol,
    AntId:    ant.ID,
    Mode:     1,
    OrderId:  orderId,
    Price:    buyPrice,
    Quantity: buyQuantity,
    Version:  1,
  })

  return
}

func (r *AntRepository) Take(ant *gamblingModels.Ant, price float64) (err error) {
  var positionSide string
  var side string
  if ant.Side == 1 {
    positionSide = "LONG"
    side = "SELL"
  } else if ant.Side == 2 {
    positionSide = "SHORT"
    side = "BUY"
  }

  log.Println("side", side, price)

  position, err := r.PositionRepository.Get(ant.Symbol, ant.Side)
  if err != nil {
    return err
  }

  if position.EntryQuantity == 0 {
    timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    if position.Timestamp > timestamp {
      return errors.New("waiting for more time")
    }
    if position.Timestamp > ant.Timestamp+9e8 {
      r.Close(ant)
      r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, 1))
      r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, 2))
    }
    return errors.New(fmt.Sprintf("[%s] %s empty position", ant.Symbol, positionSide))
  }

  if position.Timestamp > ant.Timestamp {
    ant.Timestamp = position.Timestamp
  }

  entity, err := r.SymbolsRepository.Get(ant.Symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  if ant.PlaceQuantity <= ant.TakeQuantity {
    return errors.New(fmt.Sprintf("[%s] %s no quantity to take", ant.Symbol, positionSide))
  }

  takeQuantity := 0.0
  for _, quantity := range ant.TakeQuantities {
    takeQuantity, _ = decimal.NewFromFloat(takeQuantity).Add(decimal.NewFromFloat(quantity)).Float64()
  }
  restQuantity, _ := decimal.NewFromFloat(ant.PlaceQuantity).Sub(decimal.NewFromFloat(ant.TakeQuantity)).Float64()

  entryPrice := position.EntryPrice

  if takeQuantity != restQuantity {
    takePrice := r.GamblingRepository.TakePrice(entryPrice, ant.Side, tickSize)

    planPrice := entryPrice
    planQuantity := restQuantity
    lastProfit := 0.0

    ant.TakePrices = []float64{}
    ant.TakeQuantities = []float64{}

    for {
      plans := r.GamblingRepository.Calc(planPrice, planQuantity, ant.Side, tickSize, stepSize)
      for _, plan := range plans {
        if plan.TakeQuantity < stepSize {
          if ant.Side == 1 {
            lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          } else {
            lastProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          }
          break
        }
        if ant.Side == 1 && plan.TakePrice > takePrice {
          lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          break
        }
        if ant.Side == 2 && plan.TakePrice < takePrice {
          lastProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          break
        }
        planPrice = plan.TakePrice
        planQuantity, _ = decimal.NewFromFloat(planQuantity).Sub(decimal.NewFromFloat(plan.TakeQuantity)).Float64()

        ant.TakePrices = append(ant.TakePrices, plan.TakePrice)
        ant.TakeQuantities = append(ant.TakeQuantities, plan.TakeQuantity)
      }
      if len(plans) == 0 || lastProfit > 0 {
        break
      }
    }

    if planQuantity > 0 {
      ant.TakePrices = append(ant.TakePrices, takePrice)
      ant.TakeQuantities = append(ant.TakeQuantities, planQuantity)
    }

    if len(ant.TakeQuantities) == 0 {
      ant.TakePrices = append(ant.TakePrices, r.GamblingRepository.TakePrice(entryPrice, ant.Side, tickSize))
      ant.TakeQuantities = append(ant.TakeQuantities, restQuantity)
    }

    r.Db.Model(&ant).Where("version", ant.Version).Updates(map[string]interface{}{
      "take_prices":     datatypes.NewJSONSlice(ant.TakePrices),
      "take_quantities": datatypes.NewJSONSlice(ant.TakeQuantities),
      "version":         gorm.Expr("version + ?", 1),
    })
  }

  var trading *tradingsModels.Ant
  r.Db.Where("ant_id=? AND mode=2 AND status=?", ant.ID, 0).Take(&trading)
  if trading.ID != "" {
    return errors.New("waiting for take order")
  }

  sellPrice := ant.TakePrices[0]
  sellQuantity := ant.TakeQuantities[0]

  if ant.Side == 1 {
    if price < sellPrice {
      return errors.New(fmt.Sprintf("take price must reach %v", sellPrice))
    }
    if sellPrice < price*0.9985 {
      sellPrice = price * 0.9985
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if ant.Side == 2 {
    if price > sellPrice {
      return errors.New(fmt.Sprintf("take price can not exceed %v", sellPrice))
    }
    if sellPrice > price*1.0015 {
      sellPrice = price * 1.0015
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_TAKE, ant.Symbol, ant.Side),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(ant.Symbol, positionSide, side, sellPrice, sellQuantity)
  if err != nil {
    _, ok := err.(apiCommon.APIError)
    if ok {
      return err
    }
    r.Db.Model(&ant).Where("version", ant.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  r.Db.Create(&tradingsModels.Ant{
    ID:       xid.New().String(),
    Symbol:   ant.Symbol,
    AntId:    ant.ID,
    Mode:     2,
    OrderId:  orderId,
    Price:    sellPrice,
    Quantity: sellQuantity,
    Version:  1,
  })

  return nil
}

func (r *AntRepository) Close(ant *gamblingModels.Ant) {
  var total int64
  r.Db.Model(&tradingsModels.Ant{}).Where("ant_id = ? AND status = ?", ant.ID, 0).Count(&total)
  if total == 0 {
    return
  }

  var tradings []*tradingsModels.Ant
  r.Db.Select([]string{"id", "version", "updated_at"}).Where("ant_id=? AND status=?", ant.ID, 0).Find(&tradings)
  timestamp := time.Now().Add(-30 * time.Minute).Unix()
  for _, trading := range tradings {
    if trading.UpdatedAt.Unix() < timestamp {
      r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "status":  5,
        "version": gorm.Expr("version + ?", 1),
      })
    }
  }

  if ant.PlaceQuantity != ant.TakeQuantity {
    r.Db.Model(&ant).Where("version", ant.Version).Updates(map[string]interface{}{
      "take_prices":     datatypes.NewJSONSlice([]float64{}),
      "take_quantities": datatypes.NewJSONSlice([]float64{}),
      "take_quantity":   gorm.Expr("place_quantity"),
      "version":         gorm.Expr("version + ?", 1),
    })
  }
}

func (r *AntRepository) Pending() map[string]float64 {
  var result []*PendingInfo
  r.Db.Model(&tradingsModels.Ant{}).Select(
    "symbol",
    "sum(sell_quantity) as quantity",
  ).Where("status", 1).Group("symbol").Find(&result)
  data := make(map[string]float64)
  for _, item := range result {
    data[item.Symbol] = item.Quantity
  }
  return data
}

func (r *AntRepository) CanBuy(ant *gamblingModels.Ant, price float64) bool {
  var buyPrice float64

  var positionSide string
  if ant.Side == 1 {
    positionSide = "LONG"
  } else if ant.Side == 2 {
    positionSide = "SHORT"
  }
  val, _ := r.Rdb.Get(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, positionSide, ant.Symbol)).Result()
  if val != "" {
    buyPrice, _ = strconv.ParseFloat(val, 64)
    if ant.Side == 1 && price >= buyPrice*0.9615 {
      return false
    }
    if ant.Side == 2 && price <= buyPrice*1.0385 {
      return false
    }
  }

  isChange := false

  var tradings []*tradingsModels.Ant
  r.Db.Select([]string{"status", "price"}).Where("ant_id=? AND mode=1 AND status IN ?", ant.ID, []int{0, 1}).Find(&tradings)
  for _, trading := range tradings {
    if trading.Status == 0 {
      return false
    }
    if ant.Side == 1 && price >= trading.Price*0.9615 {
      return false
    }
    if ant.Side == 2 && price <= trading.Price*1.0385 {
      return false
    }
    if buyPrice == 0 {
      buyPrice = trading.Price
      isChange = true
    } else {
      if ant.Side == 1 && buyPrice > trading.Price {
        buyPrice = trading.Price
        isChange = true
      }
      if ant.Side == 2 && buyPrice < trading.Price {
        buyPrice = trading.Price
        isChange = true
      }
    }
  }

  if isChange {
    r.Rdb.Set(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, 1), buyPrice, -1)
  }

  return true
}
