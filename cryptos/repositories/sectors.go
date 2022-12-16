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

func (r *SectorsRepository) Get(slug string) (*models.Sector, error) {
	var entity *models.Sector
	result := r.Db.Where("slug", slug).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SectorsRepository) Add(name string, slug string) error {
	var entity *models.Sector
	result := r.Db.Where("slug", slug).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entity = &models.Sector{
			ID:   xid.New().String(),
			Name: name,
			Slug: slug,
		}
		r.Db.Create(&entity)
	} else {
		entity.Name = name
		r.Db.Model(&models.Sector{ID: entity.ID}).Updates(entity)
	}
	return nil
}
