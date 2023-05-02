package tradings

import (
  "context"
  "errors"
  "log"
  "time"

  "github.com/adshao/go-binance/v2/common"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/tradings"
)

type ScalpingRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository SpotSymbolsRepository
  OrdersRepository  OrdersRepository
  AccountRepository AccountRepository
}

func (r *ScalpingRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&models.Scalping{}).Where("status", []int{0, 1, 2}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *ScalpingRepository) Flush(symbol string) error {
  price, err := r.SymbolsRepository.Price(symbol)
  if err != nil {
    return err
  }
  err = r.Take(symbol, price)
  if err != nil {
    log.Println("take error", err)
  }

  var entities []*models.Scalping
  r.Db.Where("symbol=? AND status IN ?", symbol, []int{0, 2}).Find(&entities)
  for _, entity := range entities {
    if entity.Status == 0 {
      timestamp := entity.CreatedAt.Unix()
      if entity.BuyOrderId == 0 {
        orderID := r.OrdersRepository.Lost(entity.Symbol, "BUY", entity.BuyPrice, timestamp-30)
        if orderID > 0 {
          entity.BuyOrderId = orderID
          if err := r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity).Error; err != nil {
            return err
          }
        } else {
          if timestamp > time.Now().Unix()-300 {
            r.Db.Model(&models.Scalping{ID: entity.ID}).Update("status", 1)
          }
          return nil
        }
      }
      if entity.BuyOrderId > 0 {
        status := r.OrdersRepository.Status(entity.Symbol, entity.BuyOrderId)
        if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
          r.OrdersRepository.Flush(entity.Symbol, entity.BuyOrderId)
          continue
        }
        if status == "FILLED" {
          entity.Status = 1
        } else {
          entity.Status = 4
        }
      }
      r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity)
    } else if entity.Status == 2 {
      timestamp := entity.UpdatedAt.Unix()
      if entity.SellOrderId == 0 {
        orderID := r.OrdersRepository.Lost(entity.Symbol, "SELL", entity.BuyPrice, timestamp-30)
        if orderID > 0 {
          entity.SellOrderId = orderID
          if err := r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity).Error; err != nil {
            return err
          }
        } else {
          if timestamp > time.Now().Unix()-300 {
            r.Db.Model(&models.Scalping{ID: entity.ID}).Update("status", 1)
          }
          return nil
        }
      }
      if entity.SellOrderId > 0 {
        status := r.OrdersRepository.Status(entity.Symbol, entity.SellOrderId)
        if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
          r.OrdersRepository.Flush(entity.Symbol, entity.SellOrderId)
          continue
        }
        if status == "FILLED" {
          entity.Status = 3
        } else {
          entity.Status = 5
        }
      }
      r.Db.Model(&models.Scalping{ID: entity.ID}).Updates(entity)
    }
  }

  return nil
}

func (r *ScalpingRepository) Place(plan *spotModels.Plan) error {
  buyPrice, buyQuantity, err := r.SymbolsRepository.Adjust(plan.Symbol, plan.Price, plan.Amount)
  if err != nil {
    return err
  }
  balance, _, err := r.AccountRepository.Balance(plan.Symbol)
  if err != nil {
    return err
  }
  if balance < buyPrice*buyQuantity {
    return errors.New("balance not enough")
  }
  var sellPrice float64
  if plan.Amount > 15 {
    sellPrice = buyPrice * 1.02
  } else {
    sellPrice = buyPrice * 1.015
  }
  sellPrice, sellQuantity, err := r.SymbolsRepository.Adjust(plan.Symbol, sellPrice, plan.Amount)
  if err != nil {
    return err
  }

  r.Db.Transaction(func(tx *gorm.DB) error {
    var remark string
    orderID, err := r.OrdersRepository.Create(plan.Symbol, "BUY", buyPrice, buyQuantity)
    if err != nil {
      remark = err.Error()
    }
    entity := &models.Scalping{
      ID:           xid.New().String(),
      Symbol:       plan.Symbol,
      BuyOrderId:   orderID,
      BuyPrice:     buyPrice,
      BuyQuantity:  buyQuantity,
      SellPrice:    sellPrice,
      SellQuantity: sellQuantity,
      Status:       0,
      Remark:       remark,
    }
    if err := tx.Create(&entity).Error; err != nil {
      apiError, ok := err.(common.APIError)
      if ok {
        if apiError.Code == -2010 {
          tx.Model(&spotModels.Plan{ID: plan.ID}).Updates(map[string]interface{}{
            "remark": err.Error(),
            "status": 4,
          })
          return nil
        }
      }
      return err
    }

    if err := tx.Model(&spotModels.Plan{ID: plan.ID}).Update("status", 1).Error; err != nil {
      return err
    }

    return nil
  })

  r.AccountRepository.Flush()

  return nil
}

func (r *ScalpingRepository) Take(symbol string, price float64) error {
  var scalping models.Scalping
  result := r.Db.Where("symbol=? AND status=?", symbol, 1).Order("sell_price asc").Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("empty scalping")
  }
  if price < scalping.SellPrice {
    return errors.New("price too low")
  }
  orderID, err := r.OrdersRepository.Create(symbol, "SELL", scalping.SellPrice, scalping.SellQuantity)
  if err != nil {
    apiError, ok := err.(common.APIError)
    if ok {
      if apiError.Code == -2010 {
        r.Db.Model(&models.Scalping{ID: scalping.ID}).Update("remark", err.Error())
        return err
      }
    }
    scalping.Remark = err.Error()
  }
  scalping.SellOrderId = orderID
  scalping.Status = 2
  if err := r.Db.Model(&models.Scalping{ID: scalping.ID}).Updates(scalping).Error; err != nil {
    return err
  }

  return nil
}
