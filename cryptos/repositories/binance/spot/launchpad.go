package spot

import (
  "errors"
  "time"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot"
)

type LaunchpadRepository struct {
  Db *gorm.DB
}

func (r *LaunchpadRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Launchpad{})
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

func (r *LaunchpadRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Launchpad {
  var grids []*models.Launchpad
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "amount",
    "multiple",
    "price",
    "entry_price",
    "entry_quantity",
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
  query.Order("updated_at DESC")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&grids)
  return grids
}

func (r *LaunchpadRepository) Apply(
  symbol string,
  capital float64,
  price float64,
  corePrice float64,
  issuedAt time.Time,
  expiredAt time.Time,
) error {
  var launchpad *models.Launchpad
  result := r.Db.Where("symbol=? AND status IN ?", symbol, []int{1, 3}).Take(&launchpad)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &models.Launchpad{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Capital:   capital,
      Price:     price,
      CorePrice: corePrice,
      IssuedAt:  issuedAt,
      ExpiredAt: expiredAt,
      Status:    1,
    }
    r.Db.Create(&entity)
  } else {
    if launchpad.Status == 3 {
      return errors.New("launchpad error waiting")
    }
    return errors.New("launchpad not finished")
  }

  return nil
}
