package spiders

import (
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"taoniu.local/cryptos/repositories"
	spiderRepositories "taoniu.local/cryptos/repositories/spiders"
)

type CrawlsRepository struct {
	Db                   *gorm.DB
	CurrenciesRepository *repositories.CurrenciesRepository
	SourcesRepository    *SourcesRepository
	CrawlsRepository     *spiderRepositories.CrawlsRepository
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

func (r *CrawlsRepository) Currencies() *repositories.CurrenciesRepository {
	if r.CurrenciesRepository == nil {
		r.CurrenciesRepository = &repositories.CurrenciesRepository{
			Db: r.Db,
		}
	}
	return r.CurrenciesRepository
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

func (r *CrawlsRepository) Request(currency string) error {
	source, err := r.Sources().Get()
	if err != nil {
		return err
	}
	crawlSource := &spiderRepositories.CrawlSource{
		Url:      strings.Replace(source.Url, "{}", currency, 1),
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
	for _, item := range result {
		if item["about"] == nil {
			continue
		}
		r.Currencies().SetAbout(currency, item["about"].(string))
	}

	return nil
}

func (r *CrawlsRepository) JSON(in interface{}) datatypes.JSON {
	var out datatypes.JSON
	buf, _ := json.Marshal(in)
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
