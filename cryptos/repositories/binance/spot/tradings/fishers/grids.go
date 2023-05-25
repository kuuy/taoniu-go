package fishers

import (
  "taoniu.local/cryptos/repositories/binance/spot/tradings"

  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot/tradings/fishers"
)

type GridsRepository struct {
  Db                *gorm.DB
  AccountRepository tradings.AccountRepository
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
