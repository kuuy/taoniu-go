package proxies

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis/v8"
)

type CrawlsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

type CrawlSource struct {
	Url       string
	Headers   map[string]string
	UseProxy  bool
	Timeout   int
	HtmlRules *HtmlExtractRules
}

type HtmlExtractField struct {
	Name string
	Node *HtmlExtractNode
}

type HtmlExtractNode struct {
	Selector string
	Index    int
}

type HtmlExtractRules struct {
	Container *HtmlExtractNode
	List      *HtmlExtractNode
	Fields    []*HtmlExtractField
}

type SocksProxy struct {
	Ip   net.IP
	Port int
}

func (r *CrawlsRepository) Request(source *CrawlSource) error {
	session := &net.Dialer{}
	tr := &http.Transport{
		DialContext:       session.DialContext,
		DisableKeepAlives: true,
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(source.Timeout) * time.Second,
	}

	log.Println("url", source.Url)
	req, _ := http.NewRequest("GET", source.Url, nil)
	for key, val := range source.Headers {
		req.Header.Set(key, val)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(
			fmt.Sprintf(
				"request error: status[%s] code[%d]",
				resp.Status,
				resp.StatusCode,
			),
		)
	}

	proxies, err := r.ExtractHtml(resp, source.HtmlRules)
	if err != nil {
		return err
	}
	for _, proxy := range proxies {
		log.Println("proxy", proxy)
	}

	return nil
}

func (r *CrawlsRepository) ExtractHtml(resp *http.Response, rules *HtmlExtractRules) ([]*SocksProxy, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var container = doc.Find(rules.Container.Selector).First()
	if container.Nodes == nil {
		return nil, errors.New("container not exists")
	}

	var proxies []*SocksProxy
	container.Find(rules.List.Selector).Each(func(i int, s *goquery.Selection) {
		var data = make(map[string]string)
		for _, field := range rules.Fields {
			data[field.Name] = s.Find(field.Node.Selector).Eq(field.Node.Index).Text()
		}
		ip := net.ParseIP(data["ip"])
		if ip == nil {
			return
		}
		port, err := strconv.Atoi(data["port"])
		if err != nil {
			return
		}
		proxies = append(proxies, &SocksProxy{
			Ip:   ip,
			Port: port,
		})
	})
	log.Println("proxy", len(proxies))

	return proxies, nil
}

func (*CrawlsRepository) IpToLong(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

func (*CrawlsRepository) LongToIp(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}
