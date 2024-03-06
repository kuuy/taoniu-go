package dydx

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "time"

  "github.com/go-redis/redis/v8"
)

type ServerRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

type ServerTime struct {
  Iso   string  `json:"iso"`
  Epoch float64 `json:"epoch"`
}

func (r *ServerRepository) Time() (int64, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(100) * time.Millisecond,
  }

  timestamp := time.Now().UnixMilli()

  url := "https://api.dydx.exchange/v3/time"
  req, _ := http.NewRequest("GET", url, nil)
  resp, err := httpClient.Do(req)
  if err != nil {
    return 0, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return 0, errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var result ServerTime
  json.NewDecoder(resp.Body).Decode(&result)

  isoTime, _ := time.Parse("2006-01-02T15:04:05.000Z", result.Iso)
  isoTimestamp := isoTime.UnixMilli()

  r.Rdb.HMSet(
    r.Ctx,
    "dydx:server",
    map[string]interface{}{
      "timestamp": isoTimestamp,
      "timediff":  timestamp - isoTimestamp,
    },
  )

  return isoTimestamp, nil
}
