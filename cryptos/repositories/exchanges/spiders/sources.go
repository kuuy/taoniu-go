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
	result := r.Db.First(&entity, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SourcesRepository) Get() (*spidersModels.Source, error) {
	var entity *spidersModels.Source
	result := r.Db.Where("slug", "currencies-exchanges").Take(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return entity, nil
}

func (r *SourcesRepository) Add() error {
	parentId := ""
	name := "Currencies Exchanges"
	slug := "currencies-exchanges"
	url := "https://coinmarketcap.com/rankings/exchanges/"
	source := &spiderRepositories.CrawlSource{
		Url: url,
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		},
		UseProxy: false,
		Timeout:  10,
		HtmlRules: &spiderRepositories.HtmlExtractRules{
			Container: &spiderRepositories.HtmlExtractNode{
				Selector: "table.cmc-table",
			},
			List: &spiderRepositories.HtmlExtractNode{
				Selector: "tbody tr",
			},
			Json: []*spiderRepositories.JsonExtract{
				{
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "#__NEXT_DATA__",
					},
					Rules: &spiderRepositories.JsonExtractRules{
						Container: "*.pageProps",
						List:      "exchange",
						Fields: []*spiderRepositories.JsonExtractField{
							{
								Name: "slug",
								Path: "slug",
							},
							{
								Name: "name",
								Path: "name",
							},
							{
								Name: "volume",
								Path: "totalVol24h",
							},
						},
					},
				},
			},
			Fields: []*spiderRepositories.HtmlExtractField{
				{
					Name: "name",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "a.cmc-link p",
					},
				},
				{
					Name: "slug",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "a.cmc-link",
						Attr:     "href",
					},
					Replace: []*spiderRepositories.Replace{
						{
							Pattern: `/exchanges/([^/]+)/`,
							Value:   "$1",
						},
					},
				},
				{
					Name: "volume",
					Node: &spiderRepositories.HtmlExtractNode{
						Selector: "td",
						Index:    4,
					},
				},
			},
		},
	}

	return r.Sources().Add(parentId, name, slug, source)
}
