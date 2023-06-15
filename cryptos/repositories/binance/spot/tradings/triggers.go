package tradings

import (
  "context"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  "log"

  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/tradings"
)

type TriggersRepository struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository SpotSymbolsRepository
  OrdersRepository  OrdersRepository
  AccountRepository AccountRepository
}

func (r *TriggersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&spotModels.Trigger{}).Where("status", []int{1, 3}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Place(symbol string) error {
  //var trigger spotModels.Trigger
  //result := r.Db.Where("symbol=? AND status=?", symbol, 0).Take(&trigger)
  //if errors.Is(result.Error, gorm.ErrRecordNotFound) {
  //  return errors.New("triggers empty")
  //}
  //price, err := r.SymbolsRepository.Price(symbol)
  //if err != nil {
  //  return err
  //}
  //if price > trigger.BuyPrice {
  //  return errors.New("price too high")
  //}
  //balance, _, err := r.AccountRepository.Balance(symbol)
  //if err != nil {
  //  return err
  //}
  //if balance < trigger.BuyPrice*trigger.BuyQuantity {
  //  return errors.New("balance not enough")
  //}
  //
  //orderID, err := r.OrdersRepository.Create(symbol, "BUY", trigger.BuyPrice, trigger.BuyQuantity)
  //if err != nil {
  //  apiError, ok := err.(common.APIError)
  //  if ok {
  //    if apiError.Code == -2010 {
  //      r.Db.Model(&spotModels.Trigger{ID: trigger.ID}).Updates(map[string]interface{}{
  //        "remark": err.Error(),
  //      })
  //      return nil
  //    }
  //  }
  //  trigger.Remark = err.Error()
  //} else {
  //  trigger.BuyOrderId = orderID
  //  trigger.Status = 1
  //}
  //
  //if err := r.Db.Model(&spotModels.Trigger{ID: trigger.ID}).Updates(trigger).Error; err != nil {
  //  return err
  //}
  //
  //r.AccountRepository.Flush()

  return nil
}

func (r *TriggersRepository) Flush(symbol string) error {
  price, err := r.SymbolsRepository.Price(symbol)
  if err != nil {
    return err
  }
  err = r.Take(symbol, price)
  if err != nil {
    log.Println("take error", err)
  }

  //var entities []*spotModels.Trigger
  //r.Db.Where("symbol=? AND status IN ?", symbol, []int{0, 2}).Find(&entities)
  //for _, entity := range entities {
  //  if entity.Status == 0 {
  //    timestamp := entity.CreatedAt.Unix()
  //    if entity.BuyOrderId == 0 {
  //      orderID := r.OrdersRepository.Lost(entity.Symbol, "BUY", entity.BuyPrice, timestamp-30)
  //      if orderID > 0 {
  //        entity.BuyOrderId = orderID
  //        if err := r.Db.Model(&spotModels.Trigger{ID: entity.ID}).Updates(entity).Error; err != nil {
  //          return err
  //        }
  //      } else {
  //        if timestamp > time.Now().Unix()-300 {
  //          r.Db.Model(&spotModels.Trigger{ID: entity.ID}).Update("status", 1)
  //        }
  //        return nil
  //      }
  //    }
  //    if entity.BuyOrderId > 0 {
  //      status := r.OrdersRepository.Status(entity.Symbol, entity.BuyOrderId)
  //      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
  //        r.OrdersRepository.Flush(entity.Symbol, entity.BuyOrderId)
  //        continue
  //      }
  //      if status == "FILLED" {
  //        entity.Status = 1
  //      } else {
  //        entity.Status = 4
  //      }
  //    }
  //    r.Db.Model(&spotModels.Trigger{ID: entity.ID}).Updates(entity)
  //  } else if entity.Status == 2 {
  //    timestamp := entity.UpdatedAt.Unix()
  //    if entity.SellOrderId == 0 {
  //      orderID := r.OrdersRepository.Lost(entity.Symbol, "SELL", entity.BuyPrice, timestamp-30)
  //      if orderID > 0 {
  //        entity.SellOrderId = orderID
  //        if err := r.Db.Model(&spotModels.Trigger{ID: entity.ID}).Updates(entity).Error; err != nil {
  //          return err
  //        }
  //      } else {
  //        if timestamp > time.Now().Unix()-300 {
  //          r.Db.Model(&spotModels.Trigger{ID: entity.ID}).Update("status", 1)
  //        }
  //        return nil
  //      }
  //    }
  //    if entity.SellOrderId > 0 {
  //      status := r.OrdersRepository.Status(entity.Symbol, entity.SellOrderId)
  //      if status == "" || status == "NEW" || status == "PARTIALLY_FILLED" {
  //        r.OrdersRepository.Flush(entity.Symbol, entity.SellOrderId)
  //        continue
  //      }
  //      if status == "FILLED" {
  //        entity.Status = 3
  //      } else {
  //        entity.Status = 5
  //      }
  //    }
  //    r.Db.Model(&spotModels.Trigger{ID: entity.ID}).Updates(entity)
  //  }
  //}

  return nil
}

func (r *TriggersRepository) Take(symbol string, price float64) error {
  //var triggers spotModels.Trigger
  //result := r.Db.Where("symbol=? AND status=?", symbol, 1).Order("sell_price asc").Take(&triggers)
  //if errors.Is(result.Error, gorm.ErrRecordNotFound) {
  //  return errors.New("empty triggers")
  //}
  //if price < triggers.SellPrice {
  //  return errors.New("price too low")
  //}
  //orderID, err := r.OrdersRepository.Apply(symbol, "SELL", triggers.SellPrice, triggers.SellQuantity)
  //if err != nil {
  //  apiError, ok := err.(common.APIError)
  //  if ok {
  //    if apiError.Code == -2010 {
  //      r.Db.Model(&spotModels.Trigger{ID: triggers.ID}).Update("remark", err.Error())
  //      return nil
  //    }
  //  }
  //  return err
  //}
  //
  //triggers.SellOrderId = orderID
  //triggers.Status = 2
  //if err := r.Db.Model(&spotModels.Trigger{ID: triggers.ID}).Updates(triggers).Error; err != nil {
  //  return err
  //}

  return nil
}

func (r *TriggersRepository) Pending() map[string]float64 {
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
