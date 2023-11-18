package repositories

import (
  "errors"
  "github.com/rs/xid"
  "gorm.io/gorm"
  "taoniu.local/account/common"
  "taoniu.local/account/models"
)

type UsersRepository struct {
  Db *gorm.DB
}

func (r *UsersRepository) Find(id string) (*models.User, error) {
  var entity *models.User
  result := r.Db.First(&entity, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *UsersRepository) Get(email string) *models.User {
  var entity models.User
  result := r.Db.Where(
    "email=?",
    email,
  ).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil
  }

  return &entity
}

func (r *UsersRepository) Create(email string, password string) error {
  var entity models.User
  result := r.Db.Where(
    "email=?",
    email,
  ).Take(&entity)
  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return errors.New("user already exists")
  }
  salt := common.GenerateSalt(16)
  hashedPassword := common.GeneratePassword(password, salt)

  entity = models.User{
    ID:       xid.New().String(),
    Email:    email,
    Password: hashedPassword,
    Salt:     salt,
    Status:   1,
  }
  r.Db.Create(&entity)

  return nil
}
