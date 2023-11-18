package repositories

import (
  "errors"
  "gorm.io/gorm"
  "taoniu.local/account/models"
)

type ProfileRepository struct {
  Db *gorm.DB
}

func (r *ProfileRepository) Find(uid string) (*models.Profile, error) {
  var entity *models.Profile
  result := r.Db.First(&entity, "uid=?", uid)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}
