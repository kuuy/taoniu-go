package spiders

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/spiders"
)

type SourcesRepository struct {
	Db *gorm.DB
}

func (r *SourcesRepository) Find(id string) (*models.Source, error) {
	var entity *models.Source
	result := r.Db.First(&entity, "id", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SourcesRepository) Get(short string) (*models.Source, error) {
	var entity *models.Source
	result := r.Db.Where("short", short).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SourcesRepository) Add(
	parentId string,
	name string,
	short string,
	source *CrawlSource,
) error {
	hash := sha1.Sum([]byte(source.Url))

	var entity *models.Source
	result := r.Db.Where("short", short).Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entity = &models.Source{
			ID:        xid.New().String(),
			ParentID:  parentId,
			Name:      name,
			Short:     short,
			Url:       source.Url,
			UrlSha1:   hex.EncodeToString(hash[:]),
			Headers:   r.JSONMap(source.Headers),
			UseProxy:  source.UseProxy,
			Timeout:   source.Timeout,
			HtmlRules: r.JSONMap(source.HtmlRules),
			Result:    r.JSON([]string{}),
		}
		r.Db.Create(&entity)
	} else {
		entity.ParentID = parentId
		entity.Name = name
		entity.Url = source.Url
		entity.UrlSha1 = hex.EncodeToString(hash[:])
		entity.Headers = r.JSONMap(source.Headers)
		entity.UseProxy = source.UseProxy
		entity.Timeout = source.Timeout
		entity.HtmlRules = r.JSONMap(source.HtmlRules)
		r.Db.Model(&models.Source{ID: entity.ID}).Updates(entity)
	}

	return nil
}

func (r *SourcesRepository) JSON(in interface{}) datatypes.JSON {
	buf, _ := json.Marshal(in)

	var out datatypes.JSON
	json.Unmarshal(buf, &out)
	return out
}

func (r *SourcesRepository) JSONMap(in interface{}) datatypes.JSONMap {
	buf, _ := json.Marshal(in)

	var out datatypes.JSONMap
	json.Unmarshal(buf, &out)
	return out
}
