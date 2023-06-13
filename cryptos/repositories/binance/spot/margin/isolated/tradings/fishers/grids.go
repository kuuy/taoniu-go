package fishers

import (
  "errors"
  "log"

  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  savingsModels "taoniu.local/cryptos/models/binance/savings"
  spotModels "taoniu.local/cryptos/models/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot/margin/isolated/tradings/fishers"
  "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type ProductsRepository interface {
  Get(asset string) (savingsModels.FlexibleProduct, error)
  Purchase(productId string, amount float64) (int64, error)
}

type GridsRepository struct {
  Db                 *gorm.DB
  AccountRepository  tradings.AccountRepository
  ProductsRepository ProductsRepository
}

type PendingInfo struct {
  Symbol   string
  Quantity float64
}

func (r *GridsRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Grid{})
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

func (r *GridsRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Grid {
  var grids []*models.Grid
  query := r.Db.Select([]string{
    "id",
    "symbol",
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

func (r *GridsRepository) Pending() map[string]float64 {
  var result []*PendingInfo
  r.Db.Model(&models.Grid{}).Select(
    "symbol",
    "sum(sell_quantity) as quantity",
  ).Where("status", 1).Group("symbol").Find(&result)
  data := make(map[string]float64)
  for _, item := range result {
    data[item.Symbol] = item.Quantity
  }
  return data
}

func (r *GridsRepository) Collect() error {
  data := r.Pending()
  for symbol, saveQuantity := range data {
    _, quantity, err := r.AccountRepository.Balance(symbol)
    if err != nil {
      continue
    }
    var entity *spotModels.Symbol
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
    multiple := decimal.NewFromFloat(quantity - saveQuantity).Div(minPurchaseAmount)
    takeQuantity, _ := multiple.Floor().Mul(minPurchaseAmount).Float64()
    if product.MinPurchaseAmount > takeQuantity {
      continue
    }
    _, err = r.AccountRepository.Transfer(entity.BaseAsset, symbol, "ISOLATED_MARGIN", "SPOT", takeQuantity)
    if err != nil {
      continue
    }
    _, err = r.ProductsRepository.Purchase(product.ProductId, takeQuantity)
    if err != nil {
      return err
    }
  }
  return nil
}
