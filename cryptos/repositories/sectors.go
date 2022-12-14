package repositories

import (
	"errors"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"taoniu.local/cryptos/models"
)

type SectorsRepository struct {
	Db *gorm.DB
}

func (r *SectorsRepository) Get(short string) (*models.Sector, error) {
	var entity *models.Sector
	result := r.Db.Where("short", short).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SectorsRepository) Add(name string, short string) error {
	var entity *models.Sector
	result := r.Db.Where("short", short).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entity = &models.Sector{
			ID:    xid.New().String(),
			Name:  name,
			Short: short,
		}
		r.Db.Create(&entity)
	} else {
		entity.Name = name
		r.Db.Model(&models.Sector{ID: entity.ID}).Updates(entity)
	}
	return nil
}
