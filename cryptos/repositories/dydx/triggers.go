package dydx

import (
  "errors"
  "time"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/dydx"
)

type TriggersRepository struct {
  Db *gorm.DB
}

func (r *TriggersRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&models.Trigger{}).Where("status", 1).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *TriggersRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Trigger{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status", 1)
  query.Count(&total)
  return total
}

func (r *TriggersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Trigger {
  var grids []*models.Trigger
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "capital",
    "price",
    "take_price",
    "stop_price",
    "take_order_id",
    "stop_order_id",
    "profit",
    "timestamp",
    "status",
    "created_at",
    "updated_at",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status", 1)
  query.Order("updated_at desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&grids)
  return grids
}

func (r *TriggersRepository) Apply(
  symbol string,
  side int,
  capital float64,
  price float64,
  expiredAt time.Time,
) error {
  var trigger *models.Trigger
  result := r.Db.Where("symbol = ? AND side = ? AND status = 1", symbol, side).Take(&trigger)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &models.Trigger{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Side:      side,
      Capital:   capital,
      Price:     price,
      ExpiredAt: expiredAt,
      Status:    1,
    }
    r.Db.Create(&entity)
  } else {
    return errors.New("trigger not finished")
  }

  return nil
}
