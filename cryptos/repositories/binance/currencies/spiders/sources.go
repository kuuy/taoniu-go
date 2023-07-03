package spiders

import (
  "errors"
  "gorm.io/gorm"
  spidersModels "taoniu.local/cryptos/models/spiders"
  spiderRepositories "taoniu.local/cryptos/repositories/spiders"
)

type SourcesRepository struct {
  Db                      *gorm.DB
  SpiderSourcesRepository *spiderRepositories.SourcesRepository
}

func (r *SourcesRepository) Sources() *spiderRepositories.SourcesRepository {
  if r.SpiderSourcesRepository == nil {
    r.SpiderSourcesRepository = &spiderRepositories.SourcesRepository{
      Db: r.Db,
    }
  }
  return r.SpiderSourcesRepository
}

func (r *SourcesRepository) Find(id string) (*spidersModels.Source, error) {
  var entity *spidersModels.Source
  result := r.Db.First(&entity, "id=?", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) Get() (*spidersModels.Source, error) {
  var entity *spidersModels.Source
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
  source := &spiderRepositories.CrawlSource{
    Url: url,
    Headers: map[string]string{
      "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
    },
    UseProxy: false,
    Timeout:  10,
    HtmlRules: &spiderRepositories.HtmlExtractRules{
      Json: []*spiderRepositories.JsonExtract{
        {
          Rules: &spiderRepositories.JsonExtractRules{
            Container: "data.0",
            List:      "details",
            Fields: []*spiderRepositories.JsonExtractField{
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
