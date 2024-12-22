package gambling

import (
  "errors"
  "time"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures/gambling"
)

type AntRepository struct {
  Db *gorm.DB
}

func (r *AntRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Ant{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status", 1)
  query.Count(&total)
  return total
}

func (r *AntRepository) Listings(conditions map[string]interface{}, current int, pageSize int) (result []*models.Ant) {
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "price",
    "quantity",
    "amount",
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
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&result)
  return result
}

func (r *AntRepository) Apply(
  symbol string,
  side int,
  entryPrice float64,
  entryQuantity float64,
  expiredAt time.Time,
) error {
  var scalping *models.Ant
  result := r.Db.Where("symbol = ? AND side = ? AND status = 1", symbol, side).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &models.Ant{
      ID:            xid.New().String(),
      Symbol:        symbol,
      Side:          side,
      EntryPrice:    entryPrice,
      EntryQuantity: entryQuantity,
      ExpiredAt:     expiredAt,
      Status:        1,
    }
    r.Db.Create(&entity)
  } else {
    return errors.New("scalping not finished")
  }
  return nil
}