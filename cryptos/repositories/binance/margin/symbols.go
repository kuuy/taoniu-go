package margin

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  config "taoniu.local/cryptos/config/binance/spot"
  spotModels "taoniu.local/cryptos/models/binance/spot"
)

type SymbolsRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *SymbolsRepository) Currencies() []string {
  var currencies []string
  r.Db.Model(spotModels.Symbol{}).Where("status=? AND is_margin=True", "TRADING").Distinct().Pluck("base_asset", &currencies)
  return currencies
}

func (r *SymbolsRepository) Symbols() []string {
  var symbols []string
  r.Db.Model(spotModels.Symbol{}).Select("symbol").Where("status=? AND is_margin=True", "TRADING").Find(&symbols)
  return symbols
}

func (r *SymbolsRepository) Get(symbol string) (entity *spotModels.Symbol, err error) {
  err = r.Db.Where("symbol", symbol).Take(&entity).Error
  return
}

func (r *SymbolsRepository) Filters(params datatypes.JSONMap) (tickSize float64, stepSize float64, notional float64, err error) {
  var filters []string
  filters = strings.Split(params["price"].(string), ",")
  tickSize, _ = strconv.ParseFloat(filters[2], 64)
  filters = strings.Split(params["quote"].(string), ",")
  stepSize, _ = strconv.ParseFloat(filters[2], 64)
  if _, ok := params["notional"]; !ok {
    notional = 10
  } else {
    notional, _ = strconv.ParseFloat(params["notional"].(string), 64)
  }
  return
}

func (r *SymbolsRepository) Count() error {
  var count int64
  r.Db.Model(spotModels.Symbol{}).Select("symbol").Where("status=? AND is_margin=True", "TRADING").Count(&count)
  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf("binance:symbols:count"),
    map[string]interface{}{
      "spot": count,
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
  data["slippage@1%"] = 0
  data["slippage@-1%"] = 0
  data["slippage@2%"] = 0
  data["slippage@-2%"] = 0
  data["slippage_percent@1%"] = 0
  data["slippage_percent@2%"] = 0
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
  result := r.Db.Model(&spotModels.Symbol{}).Select("depth").Where("symbol", symbol).Take(&depth)
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
    fmt.Sprintf(config.REDIS_KEY_TICKERS, symbol),
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

  timestamp := time.Now().Unix()
  price, _ := strconv.ParseFloat(data[0].(string), 64)
  lasttime, _ := strconv.ParseInt(data[1].(string), 10, 64)

  if timestamp-lasttime > 30 {
    return 0, errors.New(fmt.Sprintf("[%s] price long time not freshed", symbol))
  }

  return price, nil
}

func (r *SymbolsRepository) Adjust(symbol string, price float64, amount float64) (float64, float64, error) {
  var entity spotModels.Symbol
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
      "binance:spot:indicators:%s:%s",
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

func (r *SymbolsRepository) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
