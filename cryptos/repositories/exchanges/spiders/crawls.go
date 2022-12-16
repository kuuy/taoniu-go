package spiders

import (
	"encoding/json"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"strconv"
	spiderModels "taoniu.local/cryptos/models/spiders"

	"taoniu.local/cryptos/repositories"
	spiderRepositories "taoniu.local/cryptos/repositories/spiders"
)

type CrawlsRepository struct {
	Db                  *gorm.DB
	ExchangesRepository *repositories.ExchangesRepository
	SourcesRepository   *SourcesRepository
	CrawlsRepository    *spiderRepositories.CrawlsRepository
}

type CrawlSource struct {
	Url       string
	Headers   map[string]string
	UseProxy  bool
	Timeout   int
	HtmlRules *HtmlExtractRules
}

type HtmlExtractField struct {
	Name string           `json:"name"`
	Node *HtmlExtractNode `json:"node"`
}

type HtmlExtractNode struct {
	Selector string `json:"selector"`
	Attr     string `json:"attr"`
	Index    int    `json:"index"`
}

type HtmlExtractRules struct {
	Container *HtmlExtractNode    `json:"container"`
	List      *HtmlExtractNode    `json:"list"`
	Fields    []*HtmlExtractField `json:"fields"`
}

func (r *CrawlsRepository) Exchanges() *repositories.ExchangesRepository {
	if r.ExchangesRepository == nil {
		r.ExchangesRepository = &repositories.ExchangesRepository{
			Db: r.Db,
		}
	}
	return r.ExchangesRepository
}

func (r *CrawlsRepository) Sources() *SourcesRepository {
	if r.SourcesRepository == nil {
		r.SourcesRepository = &SourcesRepository{
			Db: r.Db,
		}
	}
	return r.SourcesRepository
}

func (r *CrawlsRepository) Crawls() *spiderRepositories.CrawlsRepository {
	if r.CrawlsRepository == nil {
		r.CrawlsRepository = &spiderRepositories.CrawlsRepository{
			Db: r.Db,
		}
	}
	return r.CrawlsRepository
}

func (r *CrawlsRepository) Request() error {
	source, err := r.Sources().Get()
	if err != nil {
		return err
	}
	crawlSource := &spiderRepositories.CrawlSource{
		Url:      source.Url,
		UseProxy: false,
		Timeout:  10,
	}
	var buf []byte
	buf, _ = source.Headers.MarshalJSON()
	json.Unmarshal(buf, &crawlSource.Headers)
	buf, _ = source.HtmlRules.MarshalJSON()
	json.Unmarshal(buf, &crawlSource.HtmlRules)

	result, err := r.Crawls().Request(crawlSource)
	if err != nil {
		return err
	}
	var slugs []string
	for _, item := range result {
		name := item["name"].(string)
		slug := item["slug"].(string)

		if name == "" {
			continue
		}

		var volume float64
		if _, ok := item["volume"].(float64); ok {
			volume = item["volume"].(float64)
		}
		if _, ok := item["volume"].(string); ok {
			volume, _ = strconv.ParseFloat(item["volume"].(string), 64)
		}
		r.Exchanges().Add(name, slug, volume)
		if !r.contains(slugs, slug) {
			slugs = append(slugs, slug)
		}
	}
	source.Result = r.JSON(slugs)
	r.Db.Model(&spiderModels.Source{ID: source.ID}).Updates(source)

	return nil
}

func (r *CrawlsRepository) JSON(in interface{}) datatypes.JSON {
	buf, _ := json.Marshal(in)

	var out datatypes.JSON
	json.Unmarshal(buf, &out)
	return out
}

func (r *CrawlsRepository) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
