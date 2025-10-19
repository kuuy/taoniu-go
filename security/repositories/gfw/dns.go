package gfw

import (
  "bufio"
  "bytes"
  "context"
  "crypto/sha1"
  "encoding/hex"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "log"
  "net"
  "net/http"
  "net/url"
  "os"
  "os/exec"
  "regexp"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/security/common"
  "taoniu.local/security/config"
  "taoniu.local/security/grpc/services"
  models "taoniu.local/security/models/gfw"
)

type DnsRepository struct {
  Db              *gorm.DB
  Rdb             *redis.Client
  Ctx             context.Context
  CrawlRepository *CrawlRepository
  Service         *services.Aes
}

func (r *DnsRepository) Crawl() *CrawlRepository {
  if r.CrawlRepository == nil {
    r.CrawlRepository = &CrawlRepository{}
  }
  return r.CrawlRepository
}

func (r *DnsRepository) Gets(domains []string) []string {
  result := make([]string, len(domains))
  for i, domain := range domains {
    ips, err := net.LookupIP(domain)
    if err != nil {
      result[i] = ""
      continue
    }
    data := make([]string, len(ips))
    for j, ip := range ips {
      data[j] = ip.String()
    }
    result[i] = strings.Join(data, ",")
  }
  return result
}

func (r *DnsRepository) Query(domains []string) ([]string, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext
  httpClient := &http.Client{
    Transport: tr,
    Timeout:   3 * time.Second,
  }

  cookie, _ := r.Rdb.Get(r.Ctx, config.REDIS_KEY_GFW_DNS_API_COOKIE).Result()

  headers := map[string]string{
    "Content-Type": "application/json",
    "User-Agent":   common.GetEnvString("GFW_DNS_API_AGENT"),
    "Cookie":       fmt.Sprintf("__test=%v", cookie),
  }

  url := "https://gfw.infinityfree.me/dns.php"
  req, _ := http.NewRequest("GET", url, nil)
  for key, val := range headers {
    req.Header.Set(key, val)
  }
  q := req.URL.Query()
  q.Add("domains", strings.Join(domains, ","))
  req.URL.RawQuery = q.Encode()
  resp, err := httpClient.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return nil, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
  }

  body, _ := io.ReadAll(resp.Body)

  var data map[string]interface{}
  json.Unmarshal(body, &data)

  if _, ok := data["success"]; !ok {
    re := regexp.MustCompile(`[abc]=toNumbers\("([^"]+)"`)
    var a, b, c string
    for _, matches := range re.FindAllStringSubmatch(string(body), -1) {
      if strings.HasPrefix(matches[0], "a=") {
        a = matches[1]
      } else if strings.HasPrefix(matches[0], "b=") {
        b = matches[1]
      } else if strings.HasPrefix(matches[0], "c=") {
        c = matches[1]
      }
    }
    if a != "" && b != "" && c != "" {
      reply, err := r.Service.Decrypt(a, b, c)
      if err != nil {
        return nil, err
      }
      r.Rdb.Set(r.Ctx, config.REDIS_KEY_GFW_DNS_API_COOKIE, reply.Result, -1)
      log.Println("result", reply)
    }
    return nil, errors.New("api request failed")
  }
  if !data["success"].(bool) {
    return nil, errors.New("api request failed")
  }

  var result []string
  for _, ips := range data["data"].([]interface{}) {
    result = append(result, ips.(string))
  }

  return result, nil
}

func (r *DnsRepository) Lookup(domain string) (result []string, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
    DialContext:       (&net.Dialer{}).DialContext,
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   3 * time.Second,
  }

  headers := map[string]string{
    "Content-Type": "application/json",
  }

  params := url.Values{}
  params.Add("addr", domain)
  params.Add("dom_dns", "true")
  params.Add("x", "19")
  params.Add("y", "10")
  payload := fmt.Sprintf("{\"domain\":\"%s\",\"dnsServer\":\"google\"}", domain)

  url := "https://www.nslookup.io/api/v1/records"
  req, _ := http.NewRequest("POST", url, bytes.NewBufferString(payload))
  for key, val := range headers {
    req.Header.Set(key, val)
  }
  resp, err := httpClient.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return nil, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
  }

  body, _ := io.ReadAll(resp.Body)

  var response *LookupResponse
  err = json.Unmarshal(body, &response)
  if err != nil {
    return
  }
  if response.Records.A.Response.Code != "NOERROR" {
    err = errors.New(fmt.Sprintf("records A response code[%v] invalid", response.Records.A.Response.Code))
    return
  }
  if response.Records.AAAA.Response.Code != "NOERROR" {
    err = errors.New(fmt.Sprintf("records AAAA response code[%v] invalid", response.Records.A.Response.Code))
    return
  }
  for _, answer := range response.Records.AAAA.Response.Answer {
    if answer.Record.RecordType != "AAAA" {
      continue
    }
    result = append(result, answer.Record.Raw)
  }
  for _, answer := range response.Records.A.Response.Answer {
    if answer.Record.RecordType != "A" {
      continue
    }
    result = append(result, answer.Record.Raw)
  }

  return
}

func (r *DnsRepository) Flush() error {
  var domains []string
  r.Db.Model(models.Dns{}).Select(
    "domain",
  ).Where(
    "status IN ?",
    []int{0, 1},
  ).Order(
    "updated_at asc",
  ).Limit(
    50,
  ).Find(
    &domains,
  )
  if len(domains) == 0 {
    return errors.New("domains empty")
  }
  result, err := r.Query(domains)
  if err != nil {
    return err
  }
  for i, ips := range result {
    if ips == "" {
      continue
    }
    domain := strings.TrimRight(domains[i], ".")
    r.Save(domain, ips, 1)
  }

  return nil
}

func (r *DnsRepository) Cache() error {
  offset := 0
  pageSize := 500

  file, err := os.Create("/tmp/gfw-zone.conf")
  if err != nil {
    return err
  }
  defer file.Close()

  file.WriteString("server:\n")

  for {
    var records []*models.Dns
    r.Db.Model(models.Dns{}).Select(
      "id",
      "domain",
      "ips",
    ).Order(
      "updated_at asc",
    ).Offset(
      offset,
    ).Limit(
      pageSize,
    ).Find(
      &records,
    )
    if len(records) == 0 {
      break
    }
    for _, record := range records {
      if record.Ips == "" {
        r.Db.Model(&models.Dns{ID: record.ID}).Update("status", 2)
        continue
      }
      ips := strings.Split(record.Ips, ",")
      for _, ip := range ips {
        if strings.Contains(ip, ":") {
          file.WriteString(fmt.Sprintf("  local-data: \"%s. IN AAAA %s\"\n", record.Domain, ip))
        } else {
          file.WriteString(fmt.Sprintf("  local-data: \"%s. IN A %s\"\n", record.Domain, ip))
          exec.Command("ipset", "add", "gfw", ip).Run()
        }
      }
    }
    offset += pageSize
  }

  return nil
}

func (r *DnsRepository) Submit(domains []string) error {
  result, err := r.Query(domains)
  if err != nil {
    return err
  }
  for i, ips := range result {
    if ips == "" {
      continue
    }
    domain := strings.TrimRight(domains[i], ".")
    r.Save(domain, ips, 1)
  }
  return nil
}

func (r *DnsRepository) Monitor() error {
  ab, err := r.Crawl().Flush()
  if err != nil {
    return err
  }

  cmd := exec.Command("tcpdump", "-nt", "-s 0", "port 53")
  stdout, err := cmd.StdoutPipe()
  cmd.Stderr = cmd.Stdout
  if err != nil {
    return err
  }
  if err = cmd.Start(); err != nil {
    return err
  }
  scanner := bufio.NewScanner(stdout)
  for scanner.Scan() {
    content := scanner.Text()
    if !strings.Contains(content, "[1au]") {
      continue
    }
    domain := strings.TrimRight(strings.Split(content, " ")[7], ".")
    if ab.ShouldBlock(domain, nil) {
      r.Save(domain, "", 0)
    }
  }

  return nil
}

func (r *DnsRepository) Save(domain string, ips string, status int) error {
  sha1 := sha1.Sum([]byte(domain))
  hash := hex.EncodeToString(sha1[:])

  var dns models.Dns
  result := r.Db.Where(
    "hash=? AND domain=?",
    hash,
    domain,
  ).Take(&dns)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    dns = models.Dns{
      ID:     xid.New().String(),
      Domain: domain,
      Hash:   hash,
      Ips:    ips,
      Status: status,
    }
    r.Db.Create(&dns)
  } else {
    dns.Ips = ips
    dns.Status = status
    r.Db.Model(&models.Dns{ID: dns.ID}).Updates(dns)
  }

  return nil
}
