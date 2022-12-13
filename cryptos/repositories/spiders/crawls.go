package spiders

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"taoniu.local/cryptos/common"
	"time"

	"gorm.io/gorm"

	"github.com/PuerkitoBio/goquery"
)

type CrawlsRepository struct {
	Db *gorm.DB
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

func (r *CrawlsRepository) Request(source *CrawlSource) ([]map[string]interface{}, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if source.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1080?timeout=2s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(source.Timeout) * time.Second,
	}

	req, _ := http.NewRequest("GET", source.Url, nil)
	for key, val := range source.Headers {
		req.Header.Set(key, val)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(
			fmt.Sprintf(
				"request error: status[%s] code[%d]",
				resp.Status,
				resp.StatusCode,
			),
		)
	}

	result, err := r.ExtractHtml(resp, source.HtmlRules)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *CrawlsRepository) ExtractHtml(resp *http.Response, rules *HtmlExtractRules) ([]map[string]interface{}, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var container = doc.Find(rules.Container.Selector).First()
	if container.Nodes == nil {
		return nil, errors.New("container not exists")
	}

	var result []map[string]interface{}
	container.Find(rules.List.Selector).Each(func(i int, s *goquery.Selection) {
		var data = make(map[string]interface{})
		for _, field := range rules.Fields {
			if field.Node.Selector != "" {
				selection := s.Find(field.Node.Selector).Eq(field.Node.Index)
				if field.Node.Attr != "" {
					data[field.Name], _ = selection.Attr(field.Node.Attr)
				} else {
					data[field.Name] = selection.Text()
				}
			} else {
				if field.Node.Attr != "" {
					data[field.Name], _ = s.Attr(field.Node.Attr)
				} else {
					data[field.Name] = s.Text()
				}
			}
		}
		result = append(result, data)
	})

	return result, nil
}

func (*CrawlsRepository) IpToLong(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

func (*CrawlsRepository) LongToIp(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}
