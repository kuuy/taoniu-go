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
	return r.Sources().Find(id)
}

func (r *SourcesRepository) Get() (*spidersModels.Source, error) {
	var entity *spidersModels.Source
	result := r.Db.Where("short", "currencies-in-sectors").Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SourcesRepository) Add() error {
	parent, err := r.Sources().Get("currencies-sectors")
	if err != nil {
		return err
	}
	name := "Currencies In Sectors"
	short := "currencies-in-sectors"
	url := "https://www.coinlore.com/crypto/{}/all"
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
					Name: "symbol",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "td span.coin-ticker",
						Index:    0,
					},
				},
				{
					Name: "price",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "td",
						Attr:     "data-fiat",
						Index:    3,
					},
				},
				{
					Name: "volume",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "td",
						Attr:     "data-sort",
						Index:    5,
					},
				},
				{
					Name: "supply",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "td",
						Attr:     "data-sort",
						Index:    6,
					},
				},
			},
		},
	}

	return r.Sources().Add(parent.ID, name, short, source)
}
