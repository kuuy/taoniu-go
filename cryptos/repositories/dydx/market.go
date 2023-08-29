package dydx

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "os"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/dydx"
)

type MarketsRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

type MarketInfo struct {
  Symbol          string  `json:"market"`
  Type            string  `json:"type"`
  BaseAsset       string  `json:"baseAsset"`
  QuoteAsset      string  `json:"quoteAsset"`
  StepSize        float64 `json:"stepSize"`
  TickSize        float64 `json:"tickSize"`
  MinOrderSize    float64 `json:"minOrderSize"`
  MaxPositionSize float64 `json:"maxPositionSize"`
  Status          string  `json:"status"`
}

func (r *MarketsRepository) Symbols() []string {
  var symbols []string
  r.Db.Model(models.Market{}).Select("symbol").Where("status", "ONLINE").Find(&symbols)
  return symbols
}

func (r *MarketsRepository) Get(symbol string) (models.Market, error) {
  var entity models.Market
  result := r.Db.Where("symbol", symbol).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return entity, result.Error
  }
  return entity, nil
}

func (r *MarketsRepository) Price(symbol string, side int) (price float64, err error) {
  fields := []string{
    "ask",
    "bid",
    "timestamp",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf("dydx.prices:%s", symbol),
    fields...,
  ).Result()

  if data[0] == nil {
    err = errors.New(fmt.Sprintf("[%s] price not exists", symbol))
    return
  }

  ask, _ := strconv.ParseFloat(data[0].(string), 64)
  bid, _ := strconv.ParseFloat(data[1].(string), 64)
  timestamp, _ := strconv.ParseInt(data[2].(string), 10, 64)
  if timestamp < time.Now().Add(-30*time.Second).UnixMilli() {
    err = errors.New(fmt.Sprintf("[%s] price long time not freshed", symbol))
  }

  if side == 1 {
    price = bid
  } else {
    price = ask
  }

  return
}

func (r *MarketsRepository) Flush() error {
  markets, err := r.Request()
  if err != nil {
    return err
  }
  for _, market := range markets {
    if market.Type != "PERPETUAL" {
      continue
    }
    if market.QuoteAsset != "USD" {
      continue
    }
    var entity models.Market
    result := r.Db.Where("symbol", market.Symbol).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = models.Market{
        ID:              xid.New().String(),
        Symbol:          market.Symbol,
        BaseAsset:       market.BaseAsset,
        QuoteAsset:      market.QuoteAsset,
        StepSize:        market.StepSize,
        TickSize:        market.TickSize,
        MinOrderSize:    market.MinOrderSize,
        MaxPositionSize: market.MaxPositionSize,
        Status:          market.Status,
      }
      r.Db.Create(&entity)
    } else {
      entity.StepSize = market.StepSize
      entity.TickSize = market.TickSize
      entity.MinOrderSize = market.MinOrderSize
      entity.MaxPositionSize = market.MaxPositionSize
      entity.Status = market.Status
      r.Db.Model(&models.Market{ID: entity.ID}).Updates(entity)
    }
  }
  return nil
}

func (r *MarketsRepository) Request() ([]*MarketInfo, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  url := fmt.Sprintf("%s/v3/markets", os.Getenv("DYDX_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
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
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["markets"]; !ok {
    return nil, errors.New("invalid response")
  }

  var markets []*MarketInfo
  for _, market := range result["markets"].(map[string]interface{}) {
    data := market.(map[string]interface{})
    marketInfo := &MarketInfo{}
    marketInfo.Symbol = data["market"].(string)
    marketInfo.Type = data["type"].(string)
    marketInfo.BaseAsset = data["baseAsset"].(string)
    marketInfo.QuoteAsset = data["quoteAsset"].(string)
    marketInfo.StepSize, _ = strconv.ParseFloat(data["stepSize"].(string), 64)
    marketInfo.TickSize, _ = strconv.ParseFloat(data["tickSize"].(string), 64)
    marketInfo.MinOrderSize, _ = strconv.ParseFloat(data["minOrderSize"].(string), 64)
    marketInfo.MaxPositionSize, _ = strconv.ParseFloat(data["maxPositionSize"].(string), 64)
    marketInfo.Status = data["status"].(string)
    markets = append(markets, marketInfo)
  }

  return markets, nil
}
