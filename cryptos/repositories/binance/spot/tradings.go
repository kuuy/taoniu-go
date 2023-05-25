package spot

import (
  "context"
  "errors"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot"
)

type TradingsRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  AccountRepository  *AccountRepository
  ProductsRepository ProductsRepository
  FishersRepository  FishersRepository
  ScalpingRepository ScalpingRepository
  TriggersRepository TriggersRepository
}

func (r *TradingsRepository) Scan() []string {
  var symbols []string
  for _, symbol := range r.ScalpingRepository.Scan() {
    if !r.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range r.TriggersRepository.Scan() {
    if !r.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  for _, symbol := range r.FishersRepository.Scan() {
    if !r.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (r *TradingsRepository) Pending() map[string]float64 {
  data := make(map[string]float64)
  for symbol, quantity := range r.ScalpingRepository.Pending() {
    if _, ok := data[symbol]; ok {
      data[symbol] += quantity
    } else {
      data[symbol] = quantity
    }
  }
  for symbol, quantity := range r.TriggersRepository.Pending() {
    if _, ok := data[symbol]; ok {
      data[symbol] += quantity
    } else {
      data[symbol] = quantity
    }
  }
  for symbol, quantity := range r.FishersRepository.Pending() {
    if _, ok := data[symbol]; ok {
      data[symbol] += quantity
    } else {
      data[symbol] = quantity
    }
  }
  return data
}

func (r *TradingsRepository) Collect() error {
  data := r.Pending()
  for symbol, pendingQuantity := range data {
    _, balanceQuantity, err := r.AccountRepository.Balance(symbol)
    if err != nil {
      continue
    }
    var entity *models.Symbol
    result := r.Db.Select([]string{"base_asset", "quote_asset"}).Where("symbol", symbol).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      continue
    }
    product, err := r.ProductsRepository.Get(entity.BaseAsset)
    if err != nil {
      log.Println("error", err)
      continue
    }
    if product.Status != "PURCHASING" {
      log.Println("error", product.Status)
      continue
    }
    minPurchaseAmount := decimal.NewFromFloat(product.MinPurchaseAmount)
    savingQuantity := decimal.NewFromFloat(balanceQuantity - pendingQuantity).Div(minPurchaseAmount)
    purchaseQuantity, _ := savingQuantity.Floor().Mul(minPurchaseAmount).Float64()
    if product.MinPurchaseAmount > purchaseQuantity {
      continue
    }
    _, err = r.ProductsRepository.Purchase(product.ProductId, purchaseQuantity)
    if err != nil {
      return err
    }
  }
  return nil
}

func (r *TradingsRepository) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
