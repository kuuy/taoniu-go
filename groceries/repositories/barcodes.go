package repositories

import (
  "errors"
  "github.com/rs/xid"
  "gorm.io/gorm"
  "taoniu.local/groceries/models"
)

type BarcodesRepository struct {
  Db *gorm.DB
}

func NewBarcodesRepository(db *gorm.DB) *BarcodesRepository {
  return &BarcodesRepository{
    Db: db,
  }
}

func (r *BarcodesRepository) Find(id string) (*models.Barcode, error) {
  var entity *models.Barcode
  result := r.Db.First(&entity, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *BarcodesRepository) Get(
  uid string,
  barcode string,
) (*models.Barcode, error) {
  var entity models.Barcode
  result := r.Db.Where(
    "uid = ? AND barcode = ?",
    uid,
    barcode,
  ).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return &entity, result.Error
  }
  return &entity, nil
}

func (r *BarcodesRepository) Create(
  uid string,
  productID string,
  barcode string,
) (id string) {
  id = xid.New().String()
  r.Db.Create(&models.Barcode{
    ID:        id,
    Uid:       uid,
    ProductID: productID,
    Barcode:   barcode,
  })
  return
}

func (r *BarcodesRepository) Update(
  id string,
  barcode string,
) (err error) {
  r.Db.Where("id = ?", id).Update("barcode", barcode)
  return
}
