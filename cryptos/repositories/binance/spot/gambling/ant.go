package gambling

import (
  "errors"
  "time"

  "github.com/rs/xid"
  "gorm.io/gorm"

  gamblingModels "taoniu.local/cryptos/models/binance/spot/gambling"
)

type AntRepository struct {
  Db *gorm.DB
}

func (r *AntRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&gamblingModels.Ant{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status", 1)
  query.Count(&total)
  return total
}

func (r *AntRepository) Listings(conditions map[string]interface{}, current int, pageSize int) (result []*gamblingModels.Ant) {
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
  entryPrice float64,
  entryQuantity float64,
  placePrices []float64,
  placeQuantities []float64,
  expiredAt time.Time,
) error {
  var scalping *gamblingModels.Ant
  result := r.Db.Where("symbol = ? AND status = 1", symbol).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &gamblingModels.Ant{
      ID:              xid.New().String(),
      Symbol:          symbol,
      Mode:            1,
      EntryPrice:      entryPrice,
      EntryQuantity:   entryQuantity,
      PlacePrices:     placePrices,
      PlaceQuantities: placeQuantities,
      TakePrices:      []float64{},
      TakeQuantities:  []float64{},
      ExpiredAt:       expiredAt,
      Status:          1,
    }
    r.Db.Create(&entity)
  } else {
    return errors.New("scalping not finished")
  }
  return nil
}
