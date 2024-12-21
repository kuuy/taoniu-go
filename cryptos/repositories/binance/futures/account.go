package futures

import (
  "context"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "net/url"
  "os"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures"
)

type AccountRepository struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Nats *nats.Conn
}

func (r *AccountRepository) Balance(asset string) (map[string]float64, error) {
  fields := []string{
    "balance",
    "free",
    "unrealized_profit",
    "margin",
    "initial_margin",
    "maint_margin",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(config.REDIS_KEY_BALANCE, asset),
    fields...,
  ).Result()
  balance := map[string]float64{}
  for i, field := range fields {
    if data[i] == nil {
      return nil, errors.New("balance not exists")
    }
    balance[field], _ = strconv.ParseFloat(data[i].(string), 64)
  }
  return balance, nil
}

func (r *AccountRepository) Flush() error {
  account, err := r.Request()
  if err != nil {
    return err
  }

  for _, coin := range account.Assets {
    balance, _ := strconv.ParseFloat(coin.Balance, 64)
    free, _ := strconv.ParseFloat(coin.Free, 64)
    unrealizedProfit, _ := strconv.ParseFloat(coin.UnrealizedProfit, 64)
    margin, _ := strconv.ParseFloat(coin.Margin, 64)
    initialMargin, _ := strconv.ParseFloat(coin.InitialMargin, 64)
    maintMargin, _ := strconv.ParseFloat(coin.MaintMargin, 64)

    if balance <= 0.0 {
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:futures:balance:%s", coin.Asset))
      continue
    }

    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf(config.REDIS_KEY_BALANCE, coin.Asset),
      map[string]interface{}{
        "balance":           balance,
        "free":              free,
        "unrealized_profit": unrealizedProfit,
        "margin":            margin,
        "initial_margin":    initialMargin,
        "maint_margin":      maintMargin,
      },
    )
  }

  var symbols []string

  timestamp := time.Now().UnixMicro()

  for _, position := range account.Positions {
    if position.Isolated || position.UpdateTime == 0 {
      continue
    }
    if position.PositionSide != "LONG" && position.PositionSide != "SHORT" {
      continue
    }

    symbols = append(symbols, position.Symbol)

    var side int
    if position.PositionSide == "LONG" {
      side = 1
    } else {
      side = 2
    }

    leverage, _ := strconv.Atoi(position.Leverage)
    entryPrice, _ := strconv.ParseFloat(position.EntryPrice, 64)
    entryQuantity, _ := strconv.ParseFloat(position.EntryQuantity, 64)
    capital, _ := strconv.ParseFloat(position.Capital, 64)
    notional, _ := strconv.ParseFloat(position.Notional, 64)

    if side == 1 && entryQuantity < 0 {
      entryQuantity = 0
    }
    if side == 2 && entryQuantity > 0 {
      entryQuantity = 0
    }
    if entryQuantity < 0 {
      entryQuantity = -entryQuantity
    }

    var entity models.Position
    result := r.Db.Where(
      "symbol=? AND side=? AND status=1",
      position.Symbol,
      side,
    ).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      if notional == 0 {
        continue
      }
      entity = models.Position{
        ID:            xid.New().String(),
        Symbol:        position.Symbol,
        Side:          side,
        Leverage:      leverage,
        Capital:       capital,
        Notional:      notional,
        EntryPrice:    entryPrice,
        EntryQuantity: entryQuantity,
        Timestamp:     timestamp,
        Status:        1,
      }
      r.Db.Create(&entity)
    } else {
      if entryPrice == entity.EntryPrice && entryQuantity == entity.EntryQuantity {
        continue
      }
      r.Db.Model(&entity).Where("version", entity.Version).Updates(map[string]interface{}{
        "leverage":       leverage,
        "capital":        capital,
        "notional":       notional,
        "entry_price":    entryPrice,
        "entry_quantity": entryQuantity,
        "timestamp":      timestamp,
        "version":        gorm.Expr("version + ?", 1),
      })
    }
  }

  if len(symbols) > 0 {
    r.Db.Where("symbol NOT IN ?", symbols).Delete(&models.Position{})
  }

  return nil
}

func (r *AccountRepository) Request() (result *AccountInfo, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  params := url.Values{}
  params.Add("timeInForce", "GTC")
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  params.Add("timestamp", fmt.Sprintf("%v", timestamp))

  var apiKey, apiSecret string
  var isTestNet bool
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    isTestNet = true
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_ACCOUNT_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_ACCOUNT_API_SECRET")
  }

  mac := hmac.New(sha256.New, []byte(apiSecret))
  _, err = mac.Write([]byte(params.Encode()))
  if err != nil {
    return nil, err
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v2/account", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v2/account", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  }

  req, _ := http.NewRequest("GET", url, nil)
  req.URL.RawQuery = params.Encode()
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", apiKey)
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

  json.NewDecoder(resp.Body).Decode(&result)

  return
}
