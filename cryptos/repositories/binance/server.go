package binance

import (
  "context"
  "encoding/json"
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
  Timestamp int64 `json:"serverTime"`
}

func (r *ServerRepository) Time() (serverTime int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   100 * time.Millisecond,
  }

  timestamp := time.Now().UnixMilli()

  ctx, cancel := context.WithTimeout(r.Ctx, 3*time.Second)
  defer cancel()

  url := "https://api.binance.com/api/v1/time"
  req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    err = fmt.Errorf(
      "request error: status[%s] code[%d]",
      resp.Status,
      resp.StatusCode,
    )
    return
  }

  var result ServerTime
  if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return
  }
  serverTime = result.Timestamp

  r.Rdb.HMSet(
    r.Ctx,
    "binance:server",
    map[string]interface{}{
      "timestamp": serverTime,
      "timediff":  timestamp - serverTime,
    },
  )

  return
}
