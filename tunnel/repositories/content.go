package repositories

import (
  "errors"
  "fmt"
  "io"
  "log"
  "net"
  "net/http"
  "strings"
  "time"

  "github.com/PuerkitoBio/goquery"
)

type ContentRepository struct {
  Method  string
  Domain  string
  Path    string
  Query   string
  Body    io.Reader
  Headers map[string]string
  Timeout int
}

func (r *ContentRepository) Process(w http.ResponseWriter) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(r.Timeout) * time.Second,
  }

  url := fmt.Sprintf("https://%s/%s", r.Domain, r.Path)

  req, _ := http.NewRequest(r.Method, url, r.Body)
  req.URL.RawQuery = r.Query
  for key, val := range r.Headers {
    req.Header.Set(key, val)
  }
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    err = errors.New("invalid response")
    log.Println("err", err)
    return
  }

  if resp.StatusCode != http.StatusOK {
    err = errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
    log.Println("err", err)
    return
  }

  contentType := resp.Header.Get("Content-Type")
  log.Println("response", resp.StatusCode, contentType)

  if strings.HasPrefix(contentType, "text/html") {
    r.ProxifyHtml(w, resp)
  } else {
    body, _ := io.ReadAll(resp.Body)
    w.Write(body)
  }

  return
}

func (r *ContentRepository) ProxifyHtml(w http.ResponseWriter, resp *http.Response) {
  doc, err := goquery.NewDocumentFromReader(resp.Body)
  if err != nil {
    return
  }

  doc.Find("[src], [href], [action]").Each(func(i int, s *goquery.Selection) {
    var attr string
    var val string
    var ok bool
    if val, ok = s.Attr("src"); ok {
      attr = "src"
    } else if val, ok = s.Attr("href"); ok {
      attr = "href"
    } else {
      attr = "action"
      val, _ = s.Attr("action")
    }

    if strings.HasPrefix(val, "data:") {
      return
    }

    if strings.Index(val, "://") != -1 {
      return
    }

    if strings.HasPrefix(val, "/") {
      val = fmt.Sprintf("/%s%s", r.Domain, val)
    } else {
      val = fmt.Sprintf("/%s/%s", r.Domain, val)
    }
    s.SetAttr(attr, val)
  })

  doc.Find("script").Each(func(i int, s *goquery.Selection) {
    if _, ok := s.Attr("src"); ok {
      return
    }
    log.Println("script", s.Text())
  })

  html, _ := doc.Html()

  w.Write([]byte(html))
}

func (r *ContentRepository) ProxifyCSS(w http.ResponseWriter, resp *http.Response) {

}
