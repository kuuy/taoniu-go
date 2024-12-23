package gambling

import (
  "context"
  "errors"
  "fmt"
  apiCommon "github.com/adshao/go-binance/v2/common"
  "github.com/rs/xid"
  "log"
  "math"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  gamblingModels "taoniu.local/cryptos/models/binance/futures/gambling"
  models "taoniu.local/cryptos/models/binance/futures/tradings/gambling"
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

func (r *AntRepository) PlaceIds() []string {
  var ids []string
  r.Db.Model(&gamblingModels.Ant{}).Where("mode=1 AND status=?", 1).Pluck("id", &ids)
  return ids
}

func (r *AntRepository) TakeIds() []string {
  var ids []string
  r.Db.Model(&gamblingModels.Ant{}).Where("place_quantity>take_quantity AND status=?", 1).Pluck("id", &ids)
  return ids
}

func (r *AntRepository) AntIds() []string {
  var ids []string
  r.Db.Model(&tradingsModels.Ant{}).Select("ant_id").Where("status IN ?", []int{0, 1}).Distinct("ant_id").Find(&ids)
  return ids
}

func (r *AntRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Ant{})
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

func (r *AntRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Ant {
  var tradings []*models.Ant
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

  var tradings []*models.Ant
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
        r.Db.Model(&ant).Updates(map[string]interface{}{
          "place_quantity": gorm.Expr("place_quantity + ?", trading.Quantity),
        })
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
        r.Db.Model(&ant).Updates(map[string]interface{}{
          "take_quantity": gorm.Expr("take_quantity + ?", trading.Quantity),
        })
        r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode))
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
        r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, trading.Mode))
      }
    }
  }

  return nil
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

    lastPrice := ant.PlanPrices[len(ant.PlanPrices)-1]
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

    for i, planPrice := range ant.PlanPrices {
      if buyPrice == 0.0 || ant.Side == 1 && buyPrice > planPrice || ant.Side == 2 && buyPrice > planPrice {
        buyPrice = planPrice
        buyQuantity = ant.PlanQuantities[i]
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

  trading := models.Ant{
    ID:       xid.New().String(),
    Symbol:   ant.Symbol,
    AntId:    ant.ID,
    Mode:     1,
    OrderId:  orderId,
    Price:    buyPrice,
    Quantity: buyQuantity,
    Version:  1,
  }
  r.Db.Create(&trading)

  return
}

func (r *AntRepository) Take(id string) (err error) {
  var ant *gamblingModels.Ant
  result := r.Db.First(&ant, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("gambling ant not found")
  }

  var positionSide string
  var side string
  if ant.Side == 1 {
    positionSide = "LONG"
    side = "SELL"
  } else if ant.Side == 2 {
    positionSide = "SHORT"
    side = "BUY"
  }
  mode := 2

  log.Println("side", side)

  //var entryPrice float64
  //var sellPrice float64
  //var trading *models.Ant

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
      r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, mode))
    }
    return errors.New(fmt.Sprintf("[%s] %s empty position", ant.Symbol, positionSide))
  }

  //entryPrice = position.EntryPrice
  if position.Timestamp > ant.Timestamp {
    ant.Timestamp = position.Timestamp
  }

  //entity, err := r.SymbolsRepository.Get(ant.Symbol)
  //if err != nil {
  //  return err
  //}

  //tickSize, _, _, err := r.SymbolsRepository.Filters(entity.Filters)
  //if err != nil {
  //  return nil
  //}

  if ant.Side == 1 {
    //result := r.Db.Where("ant_id=? AND status=?", ant.ID, 1).Take(&trading)
    //if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    //  return errors.New("empty scalping")
    //}
    //if price < trading.SellPrice {
    //  if price < entryPrice*1.0105 {
    //    return errors.New("compare with sell price too low")
    //  }
    //  timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    //  if trading.UpdatedAt.UnixMicro() > timestamp {
    //    return errors.New("waiting for more time")
    //  }
    //  sellPrice = entryPrice * 1.0105
    //} else {
    //  if entryPrice > trading.SellPrice {
    //    if price < entryPrice*1.0105 {
    //      return errors.New("compare with entry price too low")
    //    }
    //    sellPrice = entryPrice * 1.0105
    //  } else {
    //    sellPrice = trading.SellPrice
    //  }
    //}
    //if sellPrice < price*0.9985 {
    //  sellPrice = price * 0.9985
    //}
    //sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  if ant.Side == 2 {
    //result := r.Db.Where("ant_id=? AND status=?", scalping.ID, 1).Order("sell_price desc").Take(&trading)
    //if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    //  return errors.New("empty scalping")
    //}
    //if price > trading.SellPrice {
    //  if price > entryPrice*0.9895 {
    //    return errors.New("price too high")
    //  }
    //  timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    //  if trading.UpdatedAt.UnixMicro() > timestamp {
    //    return errors.New("waiting for more time")
    //  }
    //  sellPrice = entryPrice * 0.9895
    //} else {
    //  if entryPrice < trading.SellPrice {
    //    if price > entryPrice*0.9895 {
    //      return errors.New("compare with entry price too high")
    //    }
    //    sellPrice = entryPrice * 0.9895
    //  } else {
    //    sellPrice = trading.SellPrice
    //  }
    //}
    //if sellPrice > price*1.0015 {
    //  sellPrice = price * 1.0015
    //}
    //sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  //orderId, err := r.OrdersRepository.Create(trading.Symbol, positionSide, side, sellPrice, trading.SellQuantity)
  //if err != nil {
  //  _, ok := err.(apiCommon.APIError)
  //  if ok {
  //    return err
  //  }
  //  r.Db.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
  //    "remark":  err.Error(),
  //    "version": gorm.Expr("version + ?", 1),
  //  })
  //}

  //r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
  //  "sell_order_id": orderId,
  //  "status":        2,
  //  "version":       gorm.Expr("version + ?", 1),
  //})

  return nil
}

func (r *AntRepository) Close(scalping *gamblingModels.Ant) {
  var total int64
  r.Db.Model(&models.Ant{}).Where("ant_id = ? AND status IN ?", scalping.ID, []int{0, 1}).Count(&total)
  if total == 0 {
    return
  }

  var tradings []*models.Ant
  r.Db.Select([]string{"id", "version", "updated_at"}).Where("ant_id=? AND status=?", scalping.ID, 1).Find(&tradings)
  timestamp := time.Now().Add(-30 * time.Minute).Unix()
  for _, trading := range tradings {
    if trading.UpdatedAt.Unix() < timestamp {
      r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "status":  5,
        "version": gorm.Expr("version + ?", 1),
      })
    }
  }
}

func (r *AntRepository) Pending() map[string]float64 {
  var result []*PendingInfo
  r.Db.Model(&models.Ant{}).Select(
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
  r.Db.Select([]string{"status", "price"}).Where("ant_id=? AND mode=? AND status IN ?", ant.ID, 1, []int{0, 1, 3}).Find(&tradings)
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

func (r *AntRepository) CanTake(ant *gamblingModels.Ant, price float64) bool {
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

  //var tradings []*models.Trigger
  //r.Db.Select([]string{"status", "buy_price"}).Where("trigger_id=? AND status IN ?", trigger.ID, []int{0, 1, 2}).Find(&tradings)
  //for _, trading := range tradings {
  //  if trading.Status == 0 {
  //    return false
  //  }
  //  if trigger.Side == 1 && price >= trading.BuyPrice*0.9615 {
  //    return false
  //  }
  //  if trigger.Side == 2 && price <= trading.BuyPrice*1.0385 {
  //    return false
  //  }
  //  if buyPrice == 0 {
  //    buyPrice = trading.BuyPrice
  //    isChange = true
  //  } else {
  //    if trigger.Side == 1 && buyPrice > trading.BuyPrice {
  //      buyPrice = trading.BuyPrice
  //      isChange = true
  //    }
  //    if trigger.Side == 2 && buyPrice < trading.BuyPrice {
  //      buyPrice = trading.BuyPrice
  //      isChange = true
  //    }
  //  }
  //}

  if isChange {
    r.Rdb.Set(r.Ctx, fmt.Sprintf(config.REDIS_KEY_TRADINGS_GAMBLING_ANT_LAST_PRICE, positionSide, ant.Symbol, 2), buyPrice, -1)
  }

  return true
}
