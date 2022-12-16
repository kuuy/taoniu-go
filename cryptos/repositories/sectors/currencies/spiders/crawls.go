package spiders

import (
	"encoding/json"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"strconv"
	"strings"
	spiderModels "taoniu.local/cryptos/models/spiders"

	"taoniu.local/cryptos/repositories"
	spiderRepositories "taoniu.local/cryptos/repositories/spiders"
)

type CrawlsRepository struct {
	Db                   *gorm.DB
	CurrenciesRepository *repositories.CurrenciesRepository
	SectorsRepository    *repositories.SectorsRepository
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

func (r *CrawlsRepository) Sectors() *repositories.SectorsRepository {
	if r.SectorsRepository == nil {
		r.SectorsRepository = &repositories.SectorsRepository{
			Db: r.Db,
		}
	}
	return r.SectorsRepository
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
	var sectors []string
	if source.ParentID != "" {
		parent, err := r.Sources().Find(source.ParentID)
		if err != nil {
			return err
		}
		var buf []byte
		buf, _ = parent.Result.MarshalJSON()
		json.Unmarshal(buf, &sectors)
	}
	var slugs []string
	var buf []byte
	buf, _ = source.Result.MarshalJSON()
	json.Unmarshal(buf, &slugs)
	for i := 0; i < len(sectors); i++ {
		sector, err := r.Sectors().Get(sectors[i])
		if err != nil {
			continue
		}
		crawlSource := &spiderRepositories.CrawlSource{
			Url:      strings.Replace(source.Url, "{}", sectors[i], 1),
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
			continue
		}
		for _, item := range result {
			symbol := item["symbol"].(string)
			slug := item["id"].(string)
			circulatingSupply, _ := strconv.ParseFloat(item["supply"].(string), 64)
			price, _ := strconv.ParseFloat(item["price"].(string), 64)
			volume, _ := strconv.ParseFloat(item["volume"].(string), 64)
			r.Currencies().Add(symbol, sector.ID, 0, circulatingSupply, price, volume)
			if !r.contains(slugs, slug) {
				slugs = append(slugs, slug)
			}
		}

		if source.ParentID == "" {
			break
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
