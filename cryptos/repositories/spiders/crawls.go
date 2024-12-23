package spiders

import (
  "encoding/binary"
  "errors"
  "fmt"
  "io"
  "net"
  "net/http"
  "regexp"
  "time"

  "github.com/PuerkitoBio/goquery"
  "github.com/tidwall/gjson"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
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
  Name    string           `json:"name"`
  Node    *HtmlExtractNode `json:"node"`
  Replace []*Replace       `json:"replace"`
}

type HtmlExtractNode struct {
  Selector string `json:"selector"`
  Attr     string `json:"attr"`
  Index    int    `json:"index"`
}

type Replace struct {
  Pattern string `json:"pattern"`
  Value   string `json:"replace"`
}

type HtmlExtractRules struct {
  Container *HtmlExtractNode    `json:"container"`
  List      *HtmlExtractNode    `json:"list"`
  Json      []*JsonExtract      `json:"json"`
  Fields    []*HtmlExtractField `json:"fields"`
}

type JsonExtract struct {
  Node  *HtmlExtractNode  `json:"node"`
  Rules *JsonExtractRules `json:"rules"`
}

type JsonExtractField struct {
  Name  string `json:"name"`
  Path  string `json:"path"`
  Match string `json:"match"`
}

type JsonExtractRules struct {
  Container string              `json:"container"`
  List      string              `json:"list"`
  Fields    []*JsonExtractField `json:"fields"`
}

func (r *CrawlsRepository) Request(source *CrawlSource) ([]map[string]interface{}, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  if source.UseProxy {
    session := &common.ProxySession{
      Proxy: fmt.Sprintf("socks5://127.0.0.1:1080?timeout=%ds", source.Timeout),
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
  var result []map[string]interface{}

  var body []byte
  var doc *goquery.Document

  if rules.Container != nil {
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
      return nil, err
    }

    var container = doc.Find(rules.Container.Selector).First()
    if container.Nodes == nil {
      return nil, errors.New("container not exists")
    }

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
          for _, replace := range field.Replace {
            m := regexp.MustCompile(replace.Pattern)
            data[field.Name] = m.ReplaceAllString(data[field.Name].(string), replace.Value)
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
  } else {
    body, _ = io.ReadAll(resp.Body)
  }

  for _, item := range rules.Json {
    if item.Node != nil {
      doc.Find(item.Node.Selector).Each(func(i int, s *goquery.Selection) {
        var container = gjson.Get(s.Text(), item.Rules.Container)
        if container.Raw == "" {
          return
        }
        container.Get(item.Rules.List).ForEach(func(_, s gjson.Result) bool {
          var data = make(map[string]interface{})
          for _, field := range item.Rules.Fields {
            if field.Match != "" && field.Match != s.Get(field.Path).Value() {
              return false
            }
            data[field.Name] = s.Get(field.Path).Value()
          }
          result = append(result, data)
          return true
        })
      })
    } else {
      var container = gjson.Get(string(body), item.Rules.Container)
      if container.Raw == "" {
        return nil, errors.New("json parse failed")
      }
      container.Get(item.Rules.List).ForEach(func(_, s gjson.Result) bool {
        var data = make(map[string]interface{})
        for _, field := range item.Rules.Fields {
          if field.Match != "" && field.Match != s.Get(field.Path).Value() {
            return false
          }
          data[field.Name] = s.Get(field.Path).Value()
        }
        result = append(result, data)
        return true
      })
    }
  }

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
