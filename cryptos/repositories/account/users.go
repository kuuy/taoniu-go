package account

import (
	"errors"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/account"
)

type UsersRepository struct {
	Db *gorm.DB
}

type UsersError struct {
	Message string
}

func (m *UsersError) Error() string {
	return m.Message
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
		return &UsersError{"user already exists"}
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
