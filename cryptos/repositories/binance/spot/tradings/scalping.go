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
  config "taoniu.local/cryptos/config/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot"
  tradingsModels "taoniu.local/cryptos/models/binance/spot/tradings"
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

func (r *ScalpingRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&models.Scalping{}).Select("symbol").Where("status", 1).Find(&symbols)
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

func (r *ScalpingRepository) Count(conditions map[string]interface{}) (total int64) {
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
  return
}

func (r *ScalpingRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*tradingsModels.Scalping {
  var tradings []*tradingsModels.Scalping
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "scalping_id",
    "plan_id",
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

  if plan.Side != 1 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("only for long side plan")
  }

  timestamp := time.Now().UnixMilli()
  if plan.Interval == "1m" && plan.Timestamp < timestamp-900000 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "15m" && plan.Timestamp < timestamp-2700000 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "4h" && plan.Timestamp < timestamp-5400000 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "1d" && plan.Timestamp < timestamp-21600000 {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("plan has been expired")
  }

  var scalping *models.Scalping
  result = r.Db.Model(&scalping).Where("symbol = ? AND status = 1", plan.Symbol).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    r.Db.Delete(&scalpingPlan, "plan_id", planId)
    return errors.New("scalping not found")
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

  var side = "BUY"

  price, err := r.SymbolsRepository.Price(plan.Symbol)
  if err != nil {
    return
  }

  buyPrice := plan.Price
  if price < buyPrice {
    buyPrice = price
  }
  buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

  var entryPrice float64

  position, err := r.PositionRepository.Get(plan.Symbol)
  if err == nil {
    if position.EntryQuantity > 0 {
      entryPrice = position.EntryPrice
    }
  }

  if entryPrice > 0 {
    if price > entryPrice {
      r.Db.Delete(&scalpingPlan, "plan_id", planId)
      return errors.New(fmt.Sprintf("scalping [%s] price big than entry price", scalping.Symbol))
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

  if price > buyPrice {
    return errors.New(fmt.Sprintf("scalping [%s] price must reach %v", scalping.Symbol, buyPrice))
  }

  if !r.CanBuy(scalping, buyPrice) {
    return errors.New(fmt.Sprintf("scalping [%s] can not buy now", scalping.Symbol))
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
    fmt.Sprintf(config.LOCKS_TRADINGS_PLACE, scalping.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(scalping.Symbol, side, buyPrice, buyQuantity)
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

func (r *ScalpingRepository) Flush(id string) (err error) {
  var scalping *models.Scalping
  var result = r.Db.Take(&scalping, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty scalping to flush")
  }

  price, err := r.SymbolsRepository.Price(scalping.Symbol)
  if err != nil {
    return
  }
  err = r.Take(scalping, price)
  if err != nil {
    log.Println("take error", scalping.Symbol, err)
  }

  placeSide := "BUY"
  takeSide := "SELL"

  var closeTrading *tradingsModels.Scalping
  r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Take(&closeTrading)

  var tradings []*tradingsModels.Scalping
  r.Db.Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 2}).Find(&tradings)

  timestamp := time.Now().Add(-15 * time.Minute).Unix()
  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, scalping.Symbol)
  for _, trading := range tradings {
    if trading.Status == 0 {
      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      if trading.BuyOrderId == 0 {
        orderId := r.OrdersRepository.Lost(trading.Symbol, placeSide, trading.BuyQuantity, trading.UpdatedAt.Add(-120*time.Second).UnixMilli())
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
          r.Rdb.Del(r.Ctx, redisKey)
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
          err = r.Db.Transaction(func(tx *gorm.DB) (err error) {
            buyQuantity, _ := decimal.NewFromFloat(trading.BuyQuantity).Add(decimal.NewFromFloat(closeTrading.BuyQuantity)).Float64()
            buyPrice, _ := decimal.NewFromFloat(trading.BuyPrice).Mul(decimal.NewFromFloat(trading.SellQuantity)).Add(
              decimal.NewFromFloat(closeTrading.BuyPrice).Mul(decimal.NewFromFloat(closeTrading.SellQuantity)),
            ).Div(decimal.NewFromFloat(buyQuantity)).Float64()
            sellPrice := buyPrice * 1.0105
            result = r.Db.Model(&closeTrading).Where("version", closeTrading.Version).Updates(map[string]interface{}{
              "status":  5,
              "version": gorm.Expr("version + ?", 1),
            })
            if result.Error != nil {
              return result.Error
            }
            if result.RowsAffected == 0 {
              return errors.New("last trading close failed")
            }
            result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
              "buy_price":     buyPrice,
              "buy_quantity":  buyQuantity,
              "sell_price":    sellPrice,
              "sell_quantity": buyQuantity,
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
            return errors.New("order update failed")
          }
        }
        r.Rdb.Set(r.Ctx, redisKey, trading.BuyPrice, time.Hour*24)
      } else if status == "CANCELED" {
        result = r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
          "buy_order_id": 0,
          "status":       4,
          "version":      gorm.Expr("version + ?", 1),
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
        orderId := r.OrdersRepository.Lost(trading.Symbol, takeSide, trading.SellQuantity, trading.UpdatedAt.Add(-120*time.Second).UnixMilli())
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

func (r *ScalpingRepository) Take(scalping *models.Scalping, price float64) (err error) {
  var side = "SELL"
  var entryPrice float64
  var sellPrice float64
  var trading *tradingsModels.Scalping

  position, err := r.PositionRepository.Get(scalping.Symbol)
  if err != nil {
    return
  }

  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, scalping.Symbol)
  if position.EntryQuantity == 0 {
    timestamp := time.Now().Add(-15 * time.Minute).UnixMicro()
    if position.Timestamp > timestamp {
      return errors.New("waiting for more time")
    }
    if position.Timestamp > scalping.Timestamp+9e8 {
      r.Close(scalping)
      r.Rdb.Del(r.Ctx, redisKey)
    }
    return errors.New(fmt.Sprintf("[%s] empty position", scalping.Symbol))
  }

  entryPrice = position.EntryPrice
  if scalping.Timestamp < position.Timestamp {
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

  result := r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Take(&trading)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty scalping to take")
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

  balance, err := r.AccountRepository.Balance(entity.BaseAsset)
  if err != nil {
    return
  }

  if balance["free"] < trading.SellQuantity {
    err = errors.New(fmt.Sprintf("[%s] free not enough", entity.BaseAsset))
    return
  }

  mutex := common.NewMutex(
    r.Rdb,
    r.Ctx,
    fmt.Sprintf(config.LOCKS_TRADINGS_TAKE, scalping.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return
  }
  defer mutex.Unlock()

  orderId, err := r.OrdersRepository.Create(trading.Symbol, side, sellPrice, trading.SellQuantity)
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

func (r *ScalpingRepository) CanBuy(
  scalping *models.Scalping,
  price float64,
) bool {
  var buyPrice float64
  redisKey := fmt.Sprintf(config.REDIS_KEY_TRADINGS_LAST_PRICE, scalping.Symbol)
  val, _ := r.Rdb.Get(r.Ctx, redisKey).Result()
  if val != "" {
    buyPrice, _ = strconv.ParseFloat(val, 64)
    if price >= buyPrice*0.9615 {
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
