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
  result := r.Db.Where("slug", "currencies-sectors").Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) Add() error {
  parentId := ""
  name := "Currencies Sectors"
  slug := "currencies-sectors"
  url := "https://www.coinlore.com/crypto-sectors"
  source := &spiderRepositories.CrawlSource{
    Url: url,
    Headers: map[string]string{
      "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
    },
    UseProxy: false,
    Timeout:  10,
    HtmlRules: &spiderRepositories.HtmlExtractRules{
      Container: &spiderRepositories.HtmlExtractNode{
        Selector: "#coins",
      },
      List: &spiderRepositories.HtmlExtractNode{
        Selector: "tbody tr",
      },
      Fields: []*spiderRepositories.HtmlExtractField{
        {
          Name: "id",
          Node: &spiderRepositories.HtmlExtractNode{
            Attr:  "id",
            Index: 0,
          },
        },
        {
          Name: "name",
          Node: &spiderRepositories.HtmlExtractNode{
            Selector: "td.text-left > a",
            Index:    0,
          },
        },
      },
    },
  }

  return r.Sources().Add(parentId, name, slug, source)
}
