package futures

import (
  "time"

  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type PlansRepository struct {
  Db *gorm.DB
}

type PlansInfo struct {
  Symbol    string
  Side      uint32
  Price     float32
  Quantity  float32
  Amount    float32
  Timestamp time.Time
}

func (r *PlansRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Plan{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Count(&total)
  return total
}

func (r *PlansRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Plan {
  var grids []*models.Plan
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "price",
    "quantity",
    "amount",
    "created_at",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Order("updated_at desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&grids)
  return grids
}
