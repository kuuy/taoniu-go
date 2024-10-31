package spot

import (
  "errors"
  "time"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/spot"
)

type ScalpingRepository struct {
  Db *gorm.DB
}

func (r *ScalpingRepository) Scan() []string {
  var symbols []string
  r.Db.Model(&models.Scalping{}).Where("status IN ?", []int{1, 2}).Distinct().Pluck("symbol", &symbols)
  return symbols
}

func (r *ScalpingRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Scalping{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  query.Where("status", 1)
  query.Count(&total)
  return total
}

func (r *ScalpingRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Scalping {
  var grids []*models.Scalping
  query := r.Db.Select([]string{
    "id",
    "symbol",
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

func (r *ScalpingRepository) Apply(
  symbol string,
  capital float64,
  price float64,
  expiredAt time.Time,
) error {
  var scalping *models.Scalping
  result := r.Db.Where("symbol = ? AND status = 1", symbol).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &models.Scalping{
      ID:        xid.New().String(),
      Symbol:    symbol,
      Capital:   capital,
      Price:     price,
      ExpiredAt: expiredAt,
      Status:    1,
    }
    r.Db.Create(&entity)
  } else {
    return errors.New("scalping not finished")
  }
  return nil
}

func (r *ScalpingRepository) PlanIds(status int) []string {
  var ids []string
  r.Db.Model(&models.ScalpingPlan{}).Select("plan_id").Where("status", status).Find(&ids)
  return ids
}

func (r *ScalpingRepository) AddPlan(planID string) error {
  var scalping *models.ScalpingPlan
  result := r.Db.Where("plan_id", planID).Take(&scalping)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity := &models.ScalpingPlan{
      PlanId: planID,
    }
    r.Db.Create(&entity)
  }
  return nil
}
