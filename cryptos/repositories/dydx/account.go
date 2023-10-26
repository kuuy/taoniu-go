package dydx

import (
  "context"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/base64"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  uuid "github.com/satori/go.uuid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/dydx"
)

type AccountRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

type AccountInfo struct {
  Balance   float64
  Free      float64
  Positions []*PositionInfo
}

type PositionInfo struct {
  Symbol        string
  Side          string
  EntryPrice    float64
  EntryQuantity float64
}

func (r *AccountRepository) Balance() (map[string]float64, error) {
  fields := []string{
    "balance",
    "free",
  }
  data, _ := r.Rdb.HMGet(r.Ctx, "dydx:balance", fields...).Result()
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
  r.Rdb.HMSet(
    r.Ctx,
    "dydx:balance",
    map[string]interface{}{
      "balance": account.Balance,
      "free":    account.Free,
    },
  )

  var symbols []string

  for _, position := range account.Positions {
    if position.Side != "LONG" && position.Side != "SHORT" {
      continue
    }

    symbols = append(symbols, position.Symbol)

    var side int
    if position.Side == "LONG" {
      side = 1
    } else {
      side = 2
    }

    if side == 1 && position.EntryQuantity < 0 {
      position.EntryQuantity = 0
    }
    if side == 2 && position.EntryQuantity > 0 {
      position.EntryQuantity = 0
    }
    if position.EntryQuantity < 0 {
      position.EntryQuantity = -position.EntryQuantity
    }

    leverage := 10
    capital := 1000000.0
    timestamp := time.Now().UnixMilli()

    var entity models.Position
    result := r.Db.Where(
      "symbol=? AND status=1",
      position.Symbol,
    ).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = models.Position{
        ID:            xid.New().String(),
        Symbol:        position.Symbol,
        Side:          side,
        Leverage:      leverage,
        Capital:       capital,
        EntryPrice:    position.EntryPrice,
        EntryQuantity: position.EntryQuantity,
        Timestamp:     timestamp,
        Status:        1,
      }
      r.Db.Create(&entity)
    } else {
      if position.EntryPrice == entity.EntryPrice && position.EntryQuantity == entity.EntryQuantity {
        continue
      }
      r.Db.Model(&entity).Where("version", entity.Version).Updates(map[string]interface{}{
        "side":           side,
        "entry_price":    position.EntryPrice,
        "entry_quantity": position.EntryQuantity,
        "timestamp":      timestamp,
        "version":        gorm.Expr("version + ?", 1),
      })
    }
  }

  if len(symbols) > 0 {
    r.Db.Model(&models.Position{}).Where("entry_quantity > 0 AND symbol NOT IN ?", symbols).Updates(map[string]interface{}{
      "entry_quantity": 0,
      "timestamp":      time.Now().UnixMilli(),
      "version":        gorm.Expr("version + ?", 1),
    })
  }

  return nil
}

func (r *AccountRepository) Request() (account *AccountInfo, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  var namespace uuid.UUID
  err = namespace.UnmarshalText([]byte(os.Getenv("DYDX_ETHEREUM_NAMESPACE")))
  if err != nil {
    return
  }
  ethereumAddress := os.Getenv("DYDX_ETHEREUM_ADDRESS")
  userID := uuid.NewV5(namespace, strings.ToLower(ethereumAddress)).String()
  accountID := uuid.NewV5(namespace, userID+strconv.Itoa(0)).String()

  path := fmt.Sprintf("/v3/accounts/%s", accountID)

  isoTimestamp := time.Unix(0, r.Timestamp()*int64(time.Millisecond)).UTC().Format("2006-01-02T15:04:05.000Z")
  payload := fmt.Sprintf("%sGET%s", isoTimestamp, path)

  secret, _ := base64.URLEncoding.DecodeString(os.Getenv("DYDX_ACCOUNT_API_SECRET"))
  mac := hmac.New(sha256.New, secret)
  _, err = mac.Write([]byte(payload))
  if err != nil {
    return
  }
  signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

  url := fmt.Sprintf("%s%s", os.Getenv("DYDX_API_ENDPOINT"), path)
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("DYDX-SIGNATURE", signature)
  req.Header.Set("DYDX-API-KEY", os.Getenv("DYDX_ACCOUNT_API_KEY"))
  req.Header.Set("DYDX-PASSPHRASE", os.Getenv("DYDX_ACCOUNT_API_PASSPHRASE"))
  req.Header.Set("DYDX-TIMESTAMP", isoTimestamp)
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

  var result map[string]interface{}
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["account"]; !ok {
    err = errors.New("invalid response")
    return
  }

  data := result["account"].(map[string]interface{})

  log.Println("account", data)

  account = &AccountInfo{}
  account.Balance, _ = strconv.ParseFloat(data["equity"].(string), 64)
  account.Free, _ = strconv.ParseFloat(data["freeCollateral"].(string), 64)

  for _, position := range data["openPositions"].(map[string]interface{}) {
    data := position.(map[string]interface{})
    positionInfo := &PositionInfo{}
    positionInfo.Symbol = data["market"].(string)
    positionInfo.Side = data["side"].(string)
    positionInfo.EntryPrice, _ = strconv.ParseFloat(data["entryPrice"].(string), 64)
    positionInfo.EntryQuantity, _ = strconv.ParseFloat(data["size"].(string), 64)
    account.Positions = append(account.Positions, positionInfo)
  }

  return
}

func (r *AccountRepository) Timestamp() int64 {
  timestamp := time.Now().UnixMilli()
  value, err := r.Rdb.HGet(r.Ctx, "dydx:server", "timediff").Result()
  if err != nil {
    return timestamp
  }
  timediff, _ := strconv.ParseInt(value, 10, 64)
  return timestamp - timediff
}
