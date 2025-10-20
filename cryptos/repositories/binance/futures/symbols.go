package futures

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "math"
  "net"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures"
)

type ExchangeInfo struct {
  Symbols []Symbol `json:"symbols"`
}

type Symbol struct {
  Symbol                string                   `json:"symbol"`
  Pair                  string                   `json:"pair"`
  ContractType          string                   `json:"contractType"`
  DeliveryDate          int64                    `json:"deliveryDate"`
  OnboardDate           int64                    `json:"onboardDate"`
  Status                string                   `json:"status"`
  MaintMarginPercent    string                   `json:"maintMarginPercent"`
  RequiredMarginPercent string                   `json:"requiredMarginPercent"`
  PricePrecision        int                      `json:"pricePrecision"`
  QuantityPrecision     int                      `json:"quantityPrecision"`
  BaseAssetPrecision    int                      `json:"baseAssetPrecision"`
  QuotePrecision        int                      `json:"quotePrecision"`
  UnderlyingType        string                   `json:"underlyingType"`
  UnderlyingSubType     []string                 `json:"underlyingSubType"`
  SettlePlan            int64                    `json:"settlePlan"`
  TriggerProtect        string                   `json:"triggerProtect"`
  OrderType             []string                 `json:"OrderType"`
  TimeInForce           []string                 `json:"timeInForce"`
  Filters               []map[string]interface{} `json:"filters"`
  QuoteAsset            string                   `json:"quoteAsset"`
  MarginAsset           string                   `json:"marginAsset"`
  BaseAsset             string                   `json:"baseAsset"`
}

type SymbolsRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *SymbolsRepository) Symbols() []string {
  var symbols []string
  r.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
  return symbols
}

func (r *SymbolsRepository) Get(symbol string) (entity *models.Symbol, err error) {
  err = r.Db.Where("symbol", symbol).Take(&entity).Error
  return
}

func (r *SymbolsRepository) Filters(params datatypes.JSONMap) (tickSize float64, stepSize float64, notional float64, err error) {
  var values []string
  values = strings.Split(params["price"].(string), ",")
  tickSize, _ = strconv.ParseFloat(values[2], 64)
  values = strings.Split(params["quote"].(string), ",")
  stepSize, _ = strconv.ParseFloat(values[2], 64)
  if _, ok := params["notional"]; !ok {
    notional = common.GetEnvFloat64("BINANCE_FUTURES_SYMBOLS_NOTIONAL")
  } else {
    notional, _ = strconv.ParseFloat(params["notional"].(string), 64)
  }
  return
}

func (r *SymbolsRepository) Flush() (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=30s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   30 * time.Second,
  }

  url := fmt.Sprintf("%s/fapi/v1/exchangeInfo", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
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

  var response ExchangeInfo
  json.NewDecoder(resp.Body).Decode(&response)

  for _, item := range response.Symbols {
    if item.QuoteAsset != "USDT" {
      continue
    }
    if item.ContractType != "PERPETUAL" {
      continue
    }
    var filters = make(datatypes.JSONMap)
    for _, filter := range item.Filters {
      if filter["filterType"].(string) == string(binance.SymbolFilterTypePriceFilter) {
        if _, ok := filter["maxPrice"]; !ok {
          continue
        }
        if _, ok := filter["minPrice"]; !ok {
          continue
        }
        if _, ok := filter["tickSize"]; !ok {
          continue
        }
        maxPrice, _ := strconv.ParseFloat(filter["maxPrice"].(string), 64)
        minPrice, _ := strconv.ParseFloat(filter["minPrice"].(string), 64)
        tickSize, _ := strconv.ParseFloat(filter["tickSize"].(string), 64)
        filters["price"] = fmt.Sprintf(
          "%s,%s,%s",
          strconv.FormatFloat(maxPrice, 'f', -1, 64),
          strconv.FormatFloat(minPrice, 'f', -1, 64),
          strconv.FormatFloat(tickSize, 'f', -1, 64),
        )
      }
      if filter["filterType"].(string) == string(binance.SymbolFilterTypeLotSize) {
        if _, ok := filter["maxQty"]; !ok {
          continue
        }
        if _, ok := filter["minQty"]; !ok {
          continue
        }
        if _, ok := filter["stepSize"]; !ok {
          continue
        }
        maxQty, _ := strconv.ParseFloat(filter["maxQty"].(string), 64)
        minQty, _ := strconv.ParseFloat(filter["minQty"].(string), 64)
        stepSize, _ := strconv.ParseFloat(filter["stepSize"].(string), 64)
        filters["quote"] = fmt.Sprintf(
          "%s,%s,%s",
          strconv.FormatFloat(maxQty, 'f', -1, 64),
          strconv.FormatFloat(minQty, 'f', -1, 64),
          strconv.FormatFloat(stepSize, 'f', -1, 64),
        )
      }
      if filter["filterType"].(string) == string(binance.SymbolFilterTypeMinNotional) {
        if _, ok := filter["notional"]; !ok {
          continue
        }
        notional, _ := strconv.ParseFloat(filter["notional"].(string), 64)
        filters["notional"] = strconv.FormatFloat(notional, 'f', -1, 64)
      }
    }
    if _, ok := filters["price"]; !ok {
      continue
    }
    if _, ok := filters["quote"]; !ok {
      continue
    }
    var entity models.Symbol
    result := r.Db.Where("symbol", item.Symbol).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = models.Symbol{
        ID:         xid.New().String(),
        Symbol:     item.Symbol,
        BaseAsset:  item.BaseAsset,
        QuoteAsset: item.QuoteAsset,
        Filters:    filters,
        Status:     item.Status,
      }
      r.Db.Create(&entity)
    } else {
      r.Db.Model(&entity).Updates(map[string]interface{}{
        "filters": filters,
        "status":  item.Status,
      })
    }
  }

  return nil
}

func (r *SymbolsRepository) Count() error {
  var count int64
  r.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Count(&count)
  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf("binance:symbols:count"),
    map[string]interface{}{
      "futures": count,
    },
  )

  return nil
}

func (r *SymbolsRepository) Slippage(symbol string) error {
  depth, err := r.Depth(symbol)
  if err != nil {
    return err
  }
  asks := depth["asks"].([]interface{})
  bids := depth["bids"].([]interface{})
  data := make(map[string]float64)
  data["slippage@1%"] = 0.0
  data["slippage@-1%"] = 0.0
  data["slippage@2%"] = 0.0
  data["slippage@-2%"] = 0.0
  data["slippage_percent@1%"] = 0.0
  data["slippage_percent@2%"] = 0.0
  var stop1, stop2 float64
  for i, item := range asks {
    price, _ := strconv.ParseFloat(item.([]interface{})[0].(string), 64)
    volume, _ := strconv.ParseFloat(item.([]interface{})[1].(string), 64)
    if i == 0 {
      stop1 = price * 1.01
      stop2 = price * 1.02
    }
    if price <= stop1 {
      data["slippage@1%"] += volume
    }
    if price > stop2 {
      break
    }
    data["slippage@2%"] += volume
  }

  for i, item := range bids {
    price, _ := strconv.ParseFloat(item.([]interface{})[0].(string), 64)
    volume, _ := strconv.ParseFloat(item.([]interface{})[1].(string), 64)
    if i == 0 {
      stop1 = price * 0.99
      stop2 = price * 0.98
    }
    if price >= stop1 {
      data["slippage@-1%"] += volume
    }
    if price < stop2 {
      break
    }
    data["slippage@-2%"] += volume
  }

  if data["slippage@1%"]+data["slippage@-1%"] == 0.0 {
    return nil
  }
  if data["slippage@2%"]+data["slippage@-2%"] == 0.0 {
    return nil
  }

  data["slippage_percent@1%"], _ = decimal.NewFromFloat(data["slippage@1%"]).Div(decimal.NewFromFloat(data["slippage@1%"]).Add(decimal.NewFromFloat(data["slippage@-1%"]))).Round(4).Float64()
  data["slippage_percent@2%"], _ = decimal.NewFromFloat(data["slippage@2%"]).Div(decimal.NewFromFloat(data["slippage@2%"]).Add(decimal.NewFromFloat(data["slippage@-2%"]))).Round(4).Float64()

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf(config.REDIS_KEY_TICKERS, symbol),
    map[string]interface{}{
      "slippage@1%":         data["slippage@1%"],
      "slippage@-1%":        data["slippage@-1%"],
      "slippage@2%":         data["slippage@2%"],
      "slippage@-2%":        data["slippage@-2%"],
      "slippage_percent@1%": data["slippage_percent@1%"],
      "slippage_percent@2%": data["slippage_percent@2%"],
    },
  )

  return nil
}

func (r *SymbolsRepository) Depth(symbol string) (map[string]interface{}, error) {
  var depth string
  result := r.Db.Model(&models.Symbol{}).Select("depth").Where("symbol", symbol).Take(&depth)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  var out map[string]interface{}
  json.Unmarshal([]byte(depth), &out)
  if len(out) == 0 {
    return nil, errors.New("depth empty")
  }
  return out, nil
}

func (r *SymbolsRepository) Price(symbol string) (float64, error) {
  fields := []string{
    "price",
    "timestamp",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      config.REDIS_KEY_TICKERS,
      symbol,
    ),
    fields...,
  ).Result()
  if len(data) != len(fields) {
    return 0, errors.New(fmt.Sprintf("[%s] price not exists", symbol))
  }
  for i := 0; i < len(fields); i++ {
    if data[i] == nil {
      return 0, errors.New(fmt.Sprintf("[%s] price not exists", symbol))
    }
  }

  timestamp := time.Now().UnixMilli()
  price, _ := strconv.ParseFloat(data[0].(string), 64)
  lasttime, _ := strconv.ParseInt(data[1].(string), 10, 64)
  if timestamp-lasttime > 30000 {
    r.Rdb.ZAdd(r.Ctx, "binance:futures:tickers:flush", &redis.Z{
      float64(timestamp),
      symbol,
    })
    return 0, errors.New(fmt.Sprintf("[%s] price long time not freshed", symbol))
  }

  return price, nil
}

func (r *SymbolsRepository) Adjust(symbol string, price float64, amount float64) (float64, float64, error) {
  var entity models.Symbol
  result := r.Db.Select("filters").Where("symbol", symbol).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return 0, 0, result.Error
  }
  var data []string
  data = strings.Split(entity.Filters["price"].(string), ",")
  maxPrice, _ := strconv.ParseFloat(data[0], 64)
  minPrice, _ := strconv.ParseFloat(data[1], 64)
  tickSize, _ := decimal.NewFromString(data[2])

  if price > maxPrice {
    return 0, 0, errors.New("price too high")
  }
  if price < minPrice {
    price = minPrice
  }

  price, _ = decimal.NewFromFloat(price).Div(tickSize).Ceil().Mul(tickSize).Float64()

  data = strings.Split(entity.Filters["quote"].(string), ",")
  maxQty, _ := strconv.ParseFloat(data[0], 64)
  minQty, _ := strconv.ParseFloat(data[1], 64)
  stepSize, _ := decimal.NewFromString(data[2])

  quantity, _ := decimal.NewFromFloat(amount).Div(decimal.NewFromFloat(price)).Div(stepSize).Ceil().Mul(stepSize).Float64()
  if quantity > maxQty {
    return 0, 0, errors.New("quantity too high")
  }
  if quantity < minQty {
    quantity = minQty
  }

  return price, quantity, nil
}

func (r *SymbolsRepository) Filter(symbol string, price float64, amount float64) (float64, float64) {
  var entity models.Symbol
  result := r.Db.Select("filters").Where("symbol", symbol).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return 0, 0
  }
  var data []string
  data = strings.Split(entity.Filters["price"].(string), ",")
  maxPrice, _ := strconv.ParseFloat(data[0], 64)
  minPrice, _ := strconv.ParseFloat(data[1], 64)
  tickSize, _ := strconv.ParseFloat(data[2], 64)

  if price > maxPrice {
    return 0, 0
  }
  if price < minPrice {
    price = minPrice
  }
  price = math.Ceil(price/tickSize) / math.Ceil(1/tickSize)

  data = strings.Split(entity.Filters["quote"].(string), ",")
  maxQty, _ := strconv.ParseFloat(data[0], 64)
  minQty, _ := strconv.ParseFloat(data[1], 64)
  stepSize, _ := strconv.ParseFloat(data[2], 64)

  quantity := math.Ceil(amount/(price*stepSize)) / math.Ceil(1/stepSize)
  if quantity > maxQty {
    return 0, 0
  }
  if quantity < minQty {
    quantity = minQty
  }

  return price, quantity
}

func (r *SymbolsRepository) Context(symbol string) map[string]interface{} {
  day := time.Now().Format("0102")
  fields := []string{
    "r3",
    "r2",
    "r1",
    "s1",
    "s2",
    "s3",
    "profit_target",
    "stop_loss_point",
    "take_profit_price",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1d:%s:%s",
      symbol,
      day,
    ),
    fields...,
  ).Result()
  var context = make(map[string]interface{})
  for i := 0; i < len(fields); i++ {
    context[fields[i]] = data[i]
  }

  return context
}
