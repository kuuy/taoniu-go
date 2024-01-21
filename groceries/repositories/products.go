package repositories

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/groceries/models"
)

type ProductsRepository struct {
  Db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
  return &ProductsRepository{
    Db: db,
  }
}

func (r *ProductsRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Product{})
  query.Count(&total)
  return total
}

func (r *ProductsRepository) Listings(conditions map[string]interface{}, current int, pageSize int) ([]*models.Product, error) {
  var entities []*models.Product
  query := r.Db.Select(
    "id",
    "title",
    "intro",
    "price",
    "cover",
  )
  query.Order("created_at desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&entities)
  return entities, nil
}

func (r *ProductsRepository) Find(id string) (*models.Product, error) {
  var entity *models.Product
  result := r.Db.First(&entity, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *ProductsRepository) Create(
  uid string,
  title string,
  intro string,
  price float64,
  cover string,
) (id string) {
  id = xid.New().String()
  r.Db.Create(&models.Product{
    ID:    id,
    Uid:   uid,
    Title: title,
    Intro: intro,
    Price: price,
    Cover: cover,
  })
  return
}

func (r *ProductsRepository) Update(
  id string,
  title string,
  intro string,
  price float64,
  cover string,
) error {
  r.Db.Where("id = ?", id).Updates(map[string]interface{}{
    "title": title,
    "intro": intro,
    "price": price,
    "cover": cover,
  })

  return nil
}
