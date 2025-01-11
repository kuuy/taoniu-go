package tradings

import (
  "context"
  "errors"
  "fmt"
  "log"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures"
  tradingsModels "taoniu.local/cryptos/models/binance/futures/tradings"
)

type ScalpingRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  SymbolsRepository  SymbolsRepository
  AccountRepository  AccountRepository
  OrdersRepository   OrdersRepository
  PositionRepository PositionRepository
}

func (r *ScalpingRepository) Scan(side int) []string {
  var symbols []string
  r.Db.Model(&models.Scalping{}).Select("symbol").Where("side = ? AND status = ?", side, 1).Find(&symbols)
  return symbols
}

func (r *ScalpingRepository) ScalpingIds() []string {
  var ids []string
  r.Db.Model(&tradingsModels.Scalping{}).Select("scalping_id").Where("status", []int{0, 1, 2}).Distinct("scalping_id").Find(&ids)
  return ids
}

func (r *ScalpingRepository) PlanIds() []string {
  var planIds []string
  r.Db.Model(&tradingsModels.Scalping{}).Where("status", []int{0, 1, 2}).Distinct().Pluck("plan_id", &planIds)
  return planIds
}

func (r *ScalpingRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&tradingsModels.Scalping{})
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

func (r *ScalpingRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*tradingsModels.Scalping {
  var tradings []*tradingsModels.Scalping
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "scalping_id",
    "plan_id",
    "buy_price",
    "sell_price",
    "buy_quantity",
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

func (r *ScalpingRepository) Flush(id string) (err error) {
  var scalping *models.Scalping
  var result = r.Db.Take(&scalping, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty scalping to flush")
  }

  price, err := r.SymbolsRepository.Price(scalping.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(scalping, price)
  if err != nil {
    log.Println("take error", scalping.Symbol, err)
  }

  var positionSide string
  var placeSide string
  var takeSide string
  if scalping.Side == 1 {
    positionSide = "LONG"
    placeSide = "BUY"
    takeSide = "SELL"
  } else if scalping.Side == 2 {
    positionSide = "SHORT"
    placeSide = "SELL"
    takeSide = "BUY"
  }

  var closeTrading *tradingsModels.Scalping
  if scalping.Side == 1 {
    r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Take(&closeTrading)
  } else {
    r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Take(&closeTrading)
  }

  var tradings []*tradingsModels.Scalping
  r.Db.Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 2}).Find(&tradings)

  timestamp := time.Now().Add(-15 * time.Minute).Unix()

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, positionSide, scalping.Symbol)

  for _, trading := range tradings {
    if trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      if trading.BuyOrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, positionSide, placeSide, trading.BuyQuantity, trading.UpdatedAt.Add(-120*time.Second).UnixMilli())
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
        if closeTrading.ID != "" && closeTrading.CreatedAt.Unix() < trading.CreatedAt.Unix() {
          sellQuantity, _ := decimal.NewFromFloat(trading.SellQuantity).Add(decimal.NewFromFloat(closeTrading.SellQuantity)).Float64()
          buyPrice, _ := decimal.NewFromFloat(trading.BuyPrice).Mul(decimal.NewFromFloat(trading.SellQuantity)).Add(
            decimal.NewFromFloat(closeTrading.BuyPrice).Mul(decimal.NewFromFloat(closeTrading.SellQuantity)),
          ).Div(decimal.NewFromFloat(sellQuantity)).Float64()
          err = r.Db.Transaction(func(tx *gorm.DB) (err error) {
            result = r.Db.Model(&closeTrading).Where("version", closeTrading.Version).Updates(map[string]interface{}{
              "status":  9,
              "version": gorm.Expr("version + ?", 1),
            })
            if result.Error != nil {
              return result.Error
            }
            if result.RowsAffected == 0 {
              log.Println("trading", closeTrading.ID, closeTrading.Version)
              return errors.New("last trading close failed")
            }
            result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
              "buy_price":     buyPrice,
              "sell_quantity": sellQuantity,
              "status":        1,
              "version":       gorm.Expr("version + ?", 1),
            })
            if result.Error != nil {
              return result.Error
            }
            if result.RowsAffected == 0 {
              return errors.New("trading update failed")
            }
            return
          })
          if err != nil {
            return
          }
        } else {
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
        orderId := r.OrdersRepository.Lost(trading.Symbol, positionSide, takeSide, trading.SellQuantity, trading.UpdatedAt.Add(-120*time.Second).UnixMilli())
        if orderId > 0 {
          status = r.OrdersRepository.Status(trading.Symbol, orderId)
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
        if trading.SellOrderId > 0 && trading.UpdatedAt.Unix() < timestamp {
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

func (r *ScalpingRepository) Place(planId string) (err error) {
  var scalpingPlan *models.ScalpingPlan
  result := r.Db.Take(&scalpingPlan, "plan_id", planId)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("scalping plan empty")
  }

  var plan *models.Plan
  result = r.Db.Take(&plan, "id", planId)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan empty")
  }

  timestamp := time.Now().Unix()
  if plan.Interval == "1m" && plan.CreatedAt.Unix() < timestamp-900 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "15m" && plan.CreatedAt.Unix() < timestamp-2700 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "4h" && plan.CreatedAt.Unix() < timestamp-5400 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "1d" && plan.CreatedAt.Unix() < timestamp-21600 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }

  var scalping *models.Scalping
  result = r.Db.Model(&scalping).Where("symbol=? AND side=? AND status=1", plan.Symbol, plan.Side).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("scalping empty")
  }

  if plan.Side == 1 && plan.Price > scalping.Price {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan price too high")
  }

  if plan.Side == 2 && plan.Price < scalping.Price {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan price too low")
  }

  entity, err := r.SymbolsRepository.Get(plan.Symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, notional, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  var positionSide string
  var side string
  if plan.Side == 1 {
    positionSide = "LONG"
    side = "BUY"
  } else if plan.Side == 2 {
    positionSide = "SHORT"
    side = "SELL"
  }

  price, err := r.SymbolsRepository.Price(plan.Symbol)
  if err != nil {
    return
  }

  buyPrice := plan.Price
  if plan.Side == 1 {
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

  var entryPrice float64
  position, err := r.PositionRepository.Get(plan.Symbol, plan.Side)
  if err == nil {
    if position.EntryQuantity > 0 {
      entryPrice = position.EntryPrice
    }
  }

  if entryPrice > 0 {
    if scalping.Side == 1 && price > entryPrice {
      r.Db.Delete(&scalpingPlan, "plan_id", planId)
      return errors.New(fmt.Sprintf("scalping [%s] long price big than entry price", scalping.Symbol))
    }
    if scalping.Side == 2 && price < entryPrice {
      r.Db.Delete(&scalpingPlan, "plan_id", planId)
      return errors.New(fmt.Sprintf("scalping [%s] short price small than entry price", scalping.Symbol))
    }
  }

  var sellPrice float64
  if plan.Side == 1 {
    if plan.Amount > 15 {
      if plan.Interval == "1m" {
        sellPrice = buyPrice * 1.0105
      } else if plan.Interval == "15m" {
        sellPrice = buyPrice * 1.0125
      } else if plan.Interval == "4h" {
        sellPrice = buyPrice * 1.0185
      } else if plan.Interval == "1d" {
        sellPrice = buyPrice * 1.0385
      }
    } else {
      if plan.Interval == "1m" {
        sellPrice = buyPrice * 1.0085
      } else if plan.Interval == "15m" {
        sellPrice = buyPrice * 1.0105
      } else if plan.Interval == "4h" {
        sellPrice = buyPrice * 1.012
      } else if plan.Interval == "1d" {
        sellPrice = buyPrice * 1.0135
      }
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    if plan.Amount > 15 {
      if plan.Interval == "1m" {
        sellPrice = buyPrice * 0.9895
      } else if plan.Interval == "15m" {
        sellPrice = buyPrice * 0.9875
      } else if plan.Interval == "4h" {
        sellPrice = buyPrice * 0.9815
      } else if plan.Interval == "1d" {
        sellPrice = buyPrice * 0.9615
      }
    } else {
      if plan.Interval == "1m" {
        sellPrice = buyPrice * 0.9915
      } else if plan.Interval == "15m" {
        sellPrice = buyPrice * 0.9895
      } else if plan.Interval == "4h" {
        sellPrice = buyPrice * 0.988
      } else if plan.Interval == "1d" {
        sellPrice = buyPrice * 0.9865
      }
    }
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  buyQuantity, _ := decimal.NewFromFloat(notional).Div(decimal.NewFromFloat(buyPrice)).Float64()
  buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()

  if plan.Side == 1 && price > buyPrice {
    return errors.New(fmt.Sprintf("scalping [%s] %s price must reach %v", scalping.Symbol, positionSide, buyPrice))
  }

  if plan.Side == 2 && price < buyPrice {
    return errors.New(fmt.Sprintf("scalping [%s] %s price must reach %v", scalping.Symbol, positionSide, buyPrice))
  }

  if !r.CanBuy(scalping, buyPrice) {
    return errors.New(fmt.Sprintf("scalping [%s] %s can not buy now", scalping.Symbol, positionSide))
  }

  balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  if err != nil {
    return
  }

  if balance["free"] < config.SCALPING_MIN_BINANCE {
    return errors.New(fmt.Sprintf("scalping free balance must reach %v", config.SCALPING_MIN_BINANCE))
  }

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_PLACE, scalping.Symbol, scalping.Side),
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(scalping.Symbol, positionSide, side, buyPrice, buyQuantity)
  if err != nil {
    if common.IsBinanceAPIError(err) {
      return
    }
    r.Db.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
      "remark":  err.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
  }

  r.Db.Model(&scalpingPlan).Where("plan_id", planId).Update("status", 1)

  trading := &tradingsModels.Scalping{
    ID:           xid.New().String(),
    Symbol:       plan.Symbol,
    ScalpingId:   scalping.ID,
    PlanId:       plan.ID,
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

func (r *ScalpingRepository) Take(scalping *models.Scalping, price float64) (err error) {
  var positionSide string
  var side string
  if scalping.Side == 1 {
    positionSide = "LONG"
    side = "SELL"
  } else if scalping.Side == 2 {
    positionSide = "SHORT"
    side = "BUY"
  }

  var entryPrice float64
  var sellPrice float64
  var trading *tradingsModels.Scalping

  position, err := r.PositionRepository.Get(scalping.Symbol, scalping.Side)
  if err != nil {
    return
  }

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, positionSide, scalping.Symbol)
  if position.EntryQuantity == 0 {
    timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    if position.Timestamp > timestamp {
      return errors.New("waiting for more time")
    }
    if position.Timestamp > scalping.Timestamp+9e8 {
      r.Close(scalping)
      r.Rdb.Del(r.Ctx, redisKey)
    }
    return errors.New(fmt.Sprintf("[%s] %s empty position", scalping.Symbol, positionSide))
  }

  entryPrice = position.EntryPrice
  if position.Timestamp > scalping.Timestamp {
    scalping.Timestamp = position.Timestamp
  }

  entity, err := r.SymbolsRepository.Get(scalping.Symbol)
  if err != nil {
    return
  }

  tickSize, _, _, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  if scalping.Side == 1 {
    result := r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Order("sell_price asc").Take(&trading)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("empty scalping")
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
  }

  if scalping.Side == 2 {
    result := r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Order("sell_price desc").Take(&trading)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return errors.New("empty scalping")
    }
    if price > trading.SellPrice {
      if price > entryPrice*0.9895 {
        return errors.New("price too high")
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

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_TAKE, scalping.Symbol, scalping.Side),
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(trading.Symbol, positionSide, side, sellPrice, trading.SellQuantity)
  if err != nil {
    if common.IsBinanceAPIError(err) {
      return
    }
    r.Db.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
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

func (r *ScalpingRepository) Close(scalping *models.Scalping) {
  var total int64
  r.Db.Model(&tradingsModels.Scalping{}).Where("scalping_id = ? AND status IN ?", scalping.ID, []int{0, 1}).Count(&total)
  if total == 0 {
    return
  }

  var tradings []*tradingsModels.Scalping
  r.Db.Select([]string{"id", "version", "updated_at"}).Where("scalping_id=? AND status=?", scalping.ID, 1).Find(&tradings)
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

func (r *ScalpingRepository) CanBuy(
  scalping *models.Scalping,
  price float64,
) bool {
  var buyPrice float64

  var positionSide string
  if scalping.Side == 1 {
    positionSide = "LONG"
  } else if scalping.Side == 2 {
    positionSide = "SHORT"
  }

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, positionSide, scalping.Symbol)
  val, _ := r.Rdb.Get(r.Ctx, redisKey).Result()
  if val != "" {
    buyPrice, _ = strconv.ParseFloat(val, 64)
    if scalping.Side == 1 && price >= buyPrice*0.9615 {
      return false
    }
    if scalping.Side == 2 && price <= buyPrice*1.0385 {
      return false
    }
  }

  isChange := false

  var tradings []*tradingsModels.Scalping
  r.Db.Select([]string{"status", "buy_price"}).Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 1, 2}).Find(&tradings)
  for _, trading := range tradings {
    if trading.Status == 0 {
      return false
    }
    if scalping.Side == 1 && price >= trading.BuyPrice*0.9615 {
      return false
    }
    if scalping.Side == 2 && price <= trading.BuyPrice*1.0385 {
      return false
    }
    if buyPrice == 0 {
      buyPrice = trading.BuyPrice
      isChange = true
    } else {
      if scalping.Side == 1 && buyPrice > trading.BuyPrice {
        buyPrice = trading.BuyPrice
        isChange = true
      }
      if scalping.Side == 2 && buyPrice < trading.BuyPrice {
        buyPrice = trading.BuyPrice
        isChange = true
      }
    }
  }

  if isChange {
    r.Rdb.Set(r.Ctx, redisKey, buyPrice, time.Hour*24)
  }

  return true
}
