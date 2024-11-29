package futures

import (
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "os"
  "time"

  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type DepthRepository struct {
  Db       *gorm.DB
  UseProxy bool
}

func (r *DepthRepository) Flush(symbol string, limit int) error {
  depth, err := r.Request(symbol, limit)
  if err != nil {
    return err
  }
  r.Db.Model(&models.Symbol{}).Where("symbol", symbol).Update("depth", depth)
  return nil
}

func (r *DepthRepository) Request(symbol string, limit int) (map[string]interface{}, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
    DialContext:       (&net.Dialer{}).DialContext,
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  url := fmt.Sprintf("%s/fapi/v1/depth", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
  q.Add("symbol", symbol)
  q.Add("limit", fmt.Sprintf("%v", limit))
  req.URL.RawQuery = q.Encode()
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

  var result map[string]interface{}
  err = json.NewDecoder(resp.Body).Decode(&result)
  if err != nil {
    return nil, err
  }
  return result, nil
}
