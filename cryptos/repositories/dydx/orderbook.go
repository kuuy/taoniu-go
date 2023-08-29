package dydx

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "os"
  "taoniu.local/cryptos/common"
  "time"

  "github.com/go-redis/redis/v8"
)

type OrderbookRepository struct {
  Rdb      *redis.Client
  Ctx      context.Context
  UseProxy bool
}

type OrderbookResponse struct {
  Asks []OrderbookInfo `json:"asks"`
  Bids []OrderbookInfo `json:"bids"`
}

type OrderbookInfo struct {
  Price  float64 `json:"price,string"`
  Volume float64 `json:"size,string"`
}

func (r *OrderbookRepository) Flush(symbol string) error {
  response, err := r.Request(symbol)
  if err != nil {
    return err
  }

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf("dydx.prices:%s", symbol),
    map[string]interface{}{
      "ask":       response.Asks[0].Price,
      "bid":       response.Bids[0].Price,
      "timestamp": time.Now().UnixMilli(),
    },
  )

  data := make(map[string]float64)
  data["slippage@1%"] = 0
  data["slippage@-1%"] = 0
  data["slippage@2%"] = 0
  data["slippage@-2%"] = 0

  var stop1, stop2 float64
  for i, ask := range response.Asks {
    if i == 0 {
      stop1 = ask.Price * 1.01
      stop2 = ask.Price * 1.02
    }
    if ask.Price <= stop1 {
      data["slippage@1%"] += ask.Volume
    }
    if ask.Price > stop2 {
      break
    }
    data["slippage@2%"] += ask.Volume
  }

  for i, bid := range response.Bids {
    if i == 0 {
      stop1 = bid.Price * 0.99
      stop2 = bid.Price * 0.98
    }
    if bid.Price >= stop1 {
      data["slippage@-1%"] += bid.Volume
    }
    if bid.Price < stop2 {
      break
    }
    data["slippage@-2%"] += bid.Volume
  }

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf("dydx:realtime:%s", symbol),
    map[string]interface{}{
      "price":        response.Bids[0].Price,
      "slippage@1%":  data["slippage@1%"],
      "slippage@2%":  data["slippage@2%"],
      "slippage@-1%": data["slippage@-1%"],
      "slippage@-2%": data["slippage@-2%"],
    },
  )

  return nil
}

func (r *OrderbookRepository) Request(symbol string) (response *OrderbookResponse, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  if r.UseProxy {
    session := &common.ProxySession{
      Proxy: "socks5://127.0.0.1:1088?timeout=5s",
    }
    tr.DialContext = session.DialContext
  } else {
    session := &net.Dialer{}
    tr.DialContext = session.DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  url := fmt.Sprintf("%s/v3/orderbook/%s", os.Getenv("DYDX_API_ENDPOINT"), symbol)
  req, _ := http.NewRequest("GET", url, nil)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    err = errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
    return
  }

  json.NewDecoder(resp.Body).Decode(&response)

  return
}
