package spiders

import (
  "errors"
  "gorm.io/gorm"
  models "taoniu.local/cryptos/models/spiders"
  repositories "taoniu.local/cryptos/repositories/spiders"
)

type SourcesRepository struct {
  Db                      *gorm.DB
  SpiderSourcesRepository *repositories.SourcesRepository
}

func (r *SourcesRepository) Sources() *repositories.SourcesRepository {
  if r.SpiderSourcesRepository == nil {
    r.SpiderSourcesRepository = &repositories.SourcesRepository{
      Db: r.Db,
    }
  }
  return r.SpiderSourcesRepository
}

func (r *SourcesRepository) Find(id string) (*models.Source, error) {
  var entity *models.Source
  result := r.Db.Take(&entity, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) Get() (*models.Source, error) {
  var entity *models.Source
  result := r.Db.Where("slug", "binance-currencies-detail").Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) Add() error {
  parentId := ""
  name := "Binance Currencies Detail"
  slug := "binance-currencies-detail"
  url := "https://www.binance.com/bapi/composite/v1/public/marketing/tardingPair/detail?symbol={}"
  source := &repositories.CrawlSource{
    Url: url,
    Headers: map[string]string{
      "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
    },
    UseProxy: false,
    Timeout:  10,
    HtmlRules: &repositories.HtmlExtractRules{
      Json: []*repositories.JsonExtract{
        {
          Rules: &repositories.JsonExtractRules{
            Container: "data.0",
            List:      "details",
            Fields: []*repositories.JsonExtractField{
              {
                Name: "about",
                Path: "description",
              },
              {
                Name:  "language",
                Path:  "language",
                Match: "CN",
              },
            },
          },
        },
      },
    },
  }

  return r.Sources().Add(parentId, name, slug, source)
}
