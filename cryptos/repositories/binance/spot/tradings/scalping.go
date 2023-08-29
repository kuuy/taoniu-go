package tradings

import (
  "context"
  "errors"
  "log"
  "time"

  "gorm.io/gorm"

  "github.com/adshao/go-binance/v2/common"
  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"

  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/tradings"
)

type ScalpingRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository SymbolsRepository
  AccountRepository AccountRepository
  OrdersRepository  OrdersRepository
}

func (r *ScalpingRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&spotModels.Scalping{}).Select("symbol").Where("status", 1).Find(&symbols)
  return symbols
}

func (r *ScalpingRepository) ScalpingIds() []string {
  var ids []string
  r.Db.Model(&models.Scalping{}).Select("scalping_id").Where("status", []int{0, 1, 2}).Distinct().Pluck("scalping_id", &ids)
  return ids
}

func (r *ScalpingRepository) PlanIds() []string {
  var planIds []string
  r.Db.Model(&models.Scalping{}).Where("status", []int{0, 1, 2}).Distinct().Pluck("plan_id", &planIds)
  return planIds
}

func (r *ScalpingRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Scalping{})
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

func (r *ScalpingRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Scalping {
  var grids []*models.Scalping
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "buy_price",
    "buy_quantity",
    "sell_price",
    "sell_quantity",
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
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&grids)
  return grids
}

func (r *ScalpingRepository) Flush(id string) error {
  var scalping *spotModels.Scalping
  result := r.Db.First(&scalping, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("scalping empty")
  }

  price, err := r.SymbolsRepository.Price(scalping.Symbol)
  if err != nil {
    return err
  }
  err = r.Take(scalping, price)
  if err != nil {
    log.Println("take error", err)
  }

  var tradings []*models.Scalping
  r.Db.Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 2}).Find(&tradings)

  for _, trading := range tradings {
    if trading.Status == 0 {
      timestamp := trading.CreatedAt.Unix()
      if trading.BuyOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, "BUY", trading.BuyQuantity, timestamp-30)
        if orderID > 0 {
          trading.BuyOrderId = orderID
          err := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
            "buy_order_id": trading.BuyOrderId,
            "version":      gorm.Expr("version + ?", 1),
          }).Error
          if err != nil {
            return err
          }
        }
      } else {
        if timestamp < time.Now().Unix()-900 {
          r.OrdersRepository.Flush(trading.Symbol, trading.BuyOrderId)
          status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
          if status == "NEW" {
            err := r.OrdersRepository.Cancel(trading.Symbol, trading.BuyOrderId)
            if err != nil {
              apiError, ok := err.(common.APIError)
              if ok {
                err := r.ApiError(apiError, scalping)
                if err != nil {
                  return err
                }
              }
            }
            r.OrdersRepository.Flush(trading.Symbol, trading.BuyOrderId)
          }
        }
      }

      status := r.OrdersRepository.Status(trading.Symbol, trading.BuyOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(trading.Symbol, trading.BuyOrderId)
        continue
      }

      if status == "FILLED" {
        trading.Status = 1
      } else {
        trading.Status = 4
      }

      err := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "buy_order_id": trading.BuyOrderId,
        "status":       trading.Status,
        "version":      gorm.Expr("version + ?", 1),
      }).Error
      if err != nil {
        return err
      }
    }

    if trading.Status == 2 {
      timestamp := trading.UpdatedAt.Unix()
      if trading.SellOrderId == 0 {
        orderID := r.OrdersRepository.Lost(trading.Symbol, "SELL", trading.SellQuantity, timestamp-30)
        if orderID > 0 {
          trading.SellOrderId = orderID
          if err := r.Db.Model(&models.Scalping{ID: trading.ID}).Updates(trading).Error; err != nil {
            return err
          }
        }
      } else {
        if timestamp < time.Now().Unix()-900 {
          r.OrdersRepository.Flush(trading.Symbol, trading.SellOrderId)
          status := r.OrdersRepository.Status(trading.Symbol, trading.SellOrderId)
          if status == "NEW" {
            err := r.OrdersRepository.Cancel(trading.Symbol, trading.SellOrderId)
            if err != nil {
              apiError, ok := err.(common.APIError)
              if ok {
                err := r.ApiError(apiError, scalping)
                if err != nil {
                  return err
                }
              }
            }
            r.OrdersRepository.Flush(trading.Symbol, trading.SellOrderId)
          }
        }
      }

      status := r.OrdersRepository.Status(trading.Symbol, trading.SellOrderId)
      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
        r.OrdersRepository.Flush(trading.Symbol, trading.SellOrderId)
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

      err := r.Db.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
        "sell_order_id": trading.SellOrderId,
        "status":        trading.Status,
        "version":       gorm.Expr("version + ?", 1),
      }).Error
      if err != nil {
        return err
      }
    }
  }

  return nil
}

func (r *ScalpingRepository) Place(planID string) error {
  var plan *spotModels.Plan
  result := r.Db.First(&plan, "id=?", planID)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("plan empty")
  }

  if plan.Side != 1 {
    r.Db.Model(&plan).Update("status", 5)
    return errors.New("plan buy only")
  }

  if plan.Amount <= 10 {
    r.Db.Model(&plan).Update("status", 5)
    return errors.New("plan a bit risk")
  }

  timestamp := time.Now().Unix()
  if plan.Interval == "1d" && plan.CreatedAt.Unix() < timestamp-21600 {
    r.Db.Model(&plan).Update("status", 4)
    return errors.New("plan has been expired")
  }
  if plan.Interval == "1m" && plan.CreatedAt.Unix() < timestamp-900 {
    r.Db.Model(&plan).Update("status", 4)
    return errors.New("plan has been expired")
  }

  var scalping *spotModels.Scalping
  result = r.Db.Model(&scalping).Where("symbol = ? AND side = ? AND status = 1", plan.Symbol, plan.Side).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    r.Db.Model(&plan).Update("status", 5)
    return errors.New("scalping empty")
  }

  entity, err := r.SymbolsRepository.Get(plan.Symbol)
  if err != nil {
    return err
  }

  tickSize, stepSize, err := r.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  price, err := r.SymbolsRepository.Price(plan.Symbol)
  if err != nil {
    return err
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

  log.Println("ticker", tickSize, stepSize)
  //var entryPrice float64
  //
  //log.Println("account", entryPrice, err)
  //if err == nil {
  //  entryPrice = position.EntryPrice
  //  if position.Timestamp > scalping.Timestamp {
  //    scalping.Timestamp = position.Timestamp
  //    if position.EntryQuantity == 0 {
  //      err := r.Close(scalping)
  //      if err != nil {
  //        return err
  //      }
  //    }
  //    err := r.Db.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
  //      "timestamp": scalping.Timestamp,
  //      "version":   gorm.Expr("version + ?", 1),
  //    }).Error
  //    if err != nil {
  //      return err
  //    }
  //  }
  //}
  //
  //if entryPrice > 0 {
  //  if scalping.Side == 1 && price > entryPrice {
  //    return errors.New(fmt.Sprintf("[%s] %s price big than entry price", scalping.Symbol, positionSide))
  //  }
  //  if scalping.Side == 2 && price < entryPrice {
  //    return errors.New(fmt.Sprintf("[%s] %s price small than entry price", scalping.Symbol, positionSide))
  //  }
  //}
  //
  //var sellPrice float64
  //if plan.Side == 1 {
  //  if plan.Amount > 15 {
  //    if plan.Interval == "1d" {
  //      sellPrice = buyPrice * 1.035
  //    }
  //    if plan.Interval == "1m" {
  //      sellPrice = buyPrice * 1.009
  //    }
  //  } else {
  //    if plan.Interval == "1d" {
  //      sellPrice = buyPrice * 1.01
  //    }
  //    if plan.Interval == "1m" {
  //      sellPrice = buyPrice * 1.005
  //    }
  //  }
  //  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  //} else {
  //  if plan.Amount > 15 {
  //    if plan.Interval == "1d" {
  //      sellPrice = buyPrice * 0.965
  //    }
  //    if plan.Interval == "1m" {
  //      sellPrice = buyPrice * 0.991
  //    }
  //  } else {
  //    if plan.Interval == "1d" {
  //      sellPrice = buyPrice * 0.99
  //    }
  //    if plan.Interval == "1m" {
  //      sellPrice = buyPrice * 0.995
  //    }
  //  }
  //  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  //}
  //
  //buyQuantity, _ := decimal.NewFromFloat(5).Div(decimal.NewFromFloat(buyPrice)).Float64()
  //buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
  //
  //buyAmount, _ := decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
  //
  //if plan.Side == 1 && price > buyPrice {
  //  return errors.New(fmt.Sprintf("[%s] %s price must reach %v", scalping.Symbol, positionSide, buyPrice))
  //}
  //
  //if plan.Side == 2 && price < buyPrice {
  //  return errors.New(fmt.Sprintf("[%s] %s price must reach %v", scalping.Symbol, positionSide, buyPrice))
  //}
  //
  //if !r.CanBuy(scalping, buyPrice) {
  //  return errors.New(fmt.Sprintf("[%s] %s can not buy now", scalping.Symbol, positionSide))
  //}
  //
  //balance, err := r.AccountRepository.Balance(entity.QuoteAsset)
  //if err != nil {
  //  return err
  //}
  //
  //if position.ID != "" && balance["margin"] < 5 {
  //  return errors.New(fmt.Sprintf("[%s] margin must reach 5", entity.QuoteAsset))
  //} else if buyAmount < balance["margin"]*float64(position.Leverage) {
  //  return errors.New(fmt.Sprintf("[%s] margin not enough", entity.QuoteAsset))
  //}
  //
  //return r.Db.Transaction(func(tx *gorm.DB) (err error) {
  //  if position.ID != "" {
  //    err = tx.Model(&position).Where("version", position.Version).Updates(map[string]interface{}{
  //      "entry_quantity": gorm.Expr("entry_quantity + ?", buyQuantity),
  //      "version":        gorm.Expr("version + ?", 1),
  //    }).Error
  //    if err != nil {
  //      return
  //    }
  //  }
  //
  //  orderID, err := r.OrdersRepository.Create(plan.Symbol, positionSide, side, buyPrice, buyQuantity)
  //  if err != nil {
  //    apiError, ok := err.(common.APIError)
  //    if ok {
  //      err := r.ApiError(apiError, scalping)
  //      if err != nil {
  //        return err
  //      }
  //    }
  //    scalping.Remark = err.Error()
  //  }
  //
  //  err = tx.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
  //    "remark":  scalping.Remark,
  //    "version": gorm.Expr("version + ?", 1),
  //  }).Error
  //  if err != nil {
  //    return
  //  }
  //
  //  err = tx.Model(&plan).Update("status", 1).Error
  //  if err != nil {
  //    return
  //  }
  //
  //  entity := &models.Scalping{
  //    ID:           xid.New().String(),
  //    Symbol:       plan.Symbol,
  //    ScalpingID:   scalping.ID,
  //    PlanID:       plan.ID,
  //    BuyOrderId:   orderID,
  //    BuyPrice:     buyPrice,
  //    BuyQuantity:  buyQuantity,
  //    SellPrice:    sellPrice,
  //    SellQuantity: buyQuantity,
  //  }
  //  return tx.Create(&entity).Error
  //})

  return nil
}

func (r *ScalpingRepository) Take(scalping *spotModels.Scalping, price float64) error {
  //var entryPrice float64
  //var sellPrice float64
  //var trading *models.Scalping
  //
  //position, err := r.PositionRepository.Get(scalping.Symbol, scalping.Side)
  //if err == nil {
  //  entryPrice = position.EntryPrice
  //  if position.Timestamp > scalping.Timestamp {
  //    scalping.Timestamp = position.Timestamp
  //    if position.EntryQuantity == 0 {
  //      return r.Close(scalping)
  //    }
  //  }
  //}
  //
  //entity, err := r.SymbolsRepository.Get(scalping.Symbol)
  //if err != nil {
  //  return err
  //}
  //
  //tickSize, _, err := r.SymbolsRepository.Filters(entity.Filters)
  //if err != nil {
  //  return nil
  //}
  //
  //if scalping.Side == 1 {
  //  result := r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Order("sell_price asc").Take(&trading)
  //  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
  //    return errors.New("empty grid")
  //  }
  //  if price < trading.SellPrice {
  //    if price < entryPrice*1.035 {
  //      return errors.New("price too low")
  //    }
  //    sellPrice = entryPrice * 1.035
  //  } else {
  //    sellPrice = trading.SellPrice
  //  }
  //  if sellPrice < price*0.9985 {
  //    sellPrice = price * 0.9985
  //  }
  //  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  //}
  //
  //if scalping.Side == 2 {
  //  result := r.Db.Where("scalping_id=? AND status=?", scalping.ID, 1).Order("sell_price desc").Take(&trading)
  //  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
  //    return errors.New("empty grid")
  //  }
  //  if price > trading.SellPrice {
  //    if price > entryPrice*0.965 {
  //      return errors.New("price too high")
  //    }
  //    sellPrice = entryPrice * 0.965
  //  } else {
  //    sellPrice = trading.SellPrice
  //  }
  //  if sellPrice > price*1.0015 {
  //    sellPrice = price * 1.0015
  //  }
  //  sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  //}
  //
  //orderID, err := r.OrdersRepository.Create(trading.Symbol, positionSide, side, sellPrice, trading.SellQuantity)
  //if err != nil {
  //  apiError, ok := err.(common.APIError)
  //  if ok {
  //    err := r.ApiError(apiError, scalping)
  //    if err != nil {
  //      return err
  //    }
  //  }
  //  scalping.Remark = err.Error()
  //}
  //
  //return r.Db.Transaction(func(tx *gorm.DB) (err error) {
  //  err = tx.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
  //    "remark":  scalping.Remark,
  //    "version": gorm.Expr("version + ?", 1),
  //  }).Error
  //  if err != nil {
  //    return
  //  }
  //
  //  err = tx.Model(&trading).Where("version", trading.Version).Updates(map[string]interface{}{
  //    "sell_order_id": orderID,
  //    "status":        2,
  //    "version":       gorm.Expr("version + ?", 1),
  //  }).Error
  //  if err != nil {
  //    return
  //  }
  //  return nil
  //})
  return nil
}

func (r *ScalpingRepository) Close(scalping *spotModels.Scalping) error {
  var total int64
  r.Db.Model(&models.Scalping{}).Where("scalping_id = ? AND status IN ?", scalping.ID, []int{0, 1, 2}).Count(&total)
  if total == 0 {
    return nil
  }
  r.Db.Model(&models.Scalping{}).Where("scalping_id = ? AND status = 0", scalping.ID).Count(&total)
  if total > 0 {
    return nil
  }
  timestamp := time.Now().Add(-15*time.Minute).UnixNano() / int64(time.Millisecond)
  if scalping.Timestamp > timestamp {
    return nil
  }
  return r.Db.Transaction(func(tx *gorm.DB) (err error) {
    err = tx.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
      "remark":  "position not exists",
      "version": gorm.Expr("version + ?", 1),
    }).Error
    if err != nil {
      return
    }
    err = tx.Model(&models.Scalping{}).Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 1, 2}).Update("status", 5).Error
    if err != nil {
      return
    }
    return
  })
}

func (r *ScalpingRepository) ApiError(apiError common.APIError, scalping *spotModels.Scalping) error {
  if apiError.Code == -1111 || apiError.Code == -1121 || apiError.Code == -2010 || apiError.Code == -4016 {
    r.Db.Model(&scalping).Where("version", scalping.Version).Updates(map[string]interface{}{
      "remark":  apiError.Error(),
      "version": gorm.Expr("version + ?", 1),
    })
    return apiError
  }
  return nil
}

func (r *ScalpingRepository) Pending() map[string]float64 {
  var result []*PendingInfo
  r.Db.Model(&models.Scalping{}).Select(
    "symbol",
    "sum(sell_quantity) as quantity",
  ).Where("status", 1).Group("symbol").Find(&result)
  data := make(map[string]float64)
  for _, item := range result {
    data[item.Symbol] = item.Quantity
  }
  return data
}

func (r *ScalpingRepository) CanBuy(
  scalping *spotModels.Scalping,
  price float64,
) bool {
  var trading models.Scalping
  result := r.Db.Where("scalping_id=? AND status IN ?", scalping.ID, []int{0, 1, 2}).Order("buy_price asc").Take(&trading)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if trading.Status == 0 {
      return false
    }
    if price >= trading.BuyPrice*0.965 {
      return false
    }
  }
  return true
}
