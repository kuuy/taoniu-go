package repositories

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/groceries/models"
)

type StoresRepository struct {
  Db *gorm.DB
}

func NewStoresRepository(db *gorm.DB) *StoresRepository {
  return &StoresRepository{
    Db: db,
  }
}

func (r *StoresRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Store{})
  query.Count(&total)
  return total
}

func (r *StoresRepository) Listings(conditions map[string]interface{}, current int, pageSize int) ([]*models.Store, error) {
  var entities []*models.Store
  query := r.Db.Select(
    "id",
    "name",
    "logo",
  )
  query.Order("created_at desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&entities)
  return entities, nil
}

func (r *StoresRepository) Find(id string) (*models.Store, error) {
  var entity *models.Store
  result := r.Db.First(&entity, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *StoresRepository) Create(
  Uid string,
  name string,
  logo string,
) (id string) {
  id = xid.New().String()
  r.Db.Create(&models.Store{
    ID:   id,
    Uid:  Uid,
    Name: name,
    Logo: logo,
  })
  return
}

func (r *StoresRepository) Update(
  id string,
  name string,
  logo string,
) error {
  r.Db.Where("id = ?", id).Updates(map[string]interface{}{
    "name": name,
    "logo": logo,
  })
  return nil
}
