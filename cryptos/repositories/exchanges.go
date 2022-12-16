package repositories

import (
	"encoding/json"
	"errors"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"taoniu.local/cryptos/models"
)

type ExchangesRepository struct {
	Db *gorm.DB
}

func (r *ExchangesRepository) Add(
	name string,
	slug string,
	volume float64,
) error {
	var entity *models.Exchange
	result := r.Db.Where("name", name).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entity = &models.Exchange{
			ID:     xid.New().String(),
			Name:   name,
			Slug:   slug,
			Volume: volume,
		}
		r.Db.Create(&entity)
	} else {
		if entity.Volume == 0 {
			entity.Volume = volume
		}
		r.Db.Model(&models.Exchange{ID: entity.ID}).Updates(entity)
	}
	return nil
}

func (r *ExchangesRepository) JSON(in interface{}) datatypes.JSON {
	buf, _ := json.Marshal(in)

	var out datatypes.JSON
	json.Unmarshal(buf, &out)
	return out
}
