package dydx

import (
  "bytes"
  "context"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/base64"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "log"
  "net"
  "net/http"
  "net/url"
  "os"
  "strconv"
  "time"

  "github.com/go-numb/go-dydx/helpers"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/yanue/starkex"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/dydx"
)

type OrdersRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

type OrderInfo struct {
  Symbol           string
  OrderId          string
  Type             string
  PositionSide     string
  Side             string
  Price            float64
  ActivatePrice    float64
  Quantity         float64
  ExecutedQuantity float64
  OpenTime         int64
  UpdateTime       int64
  ReduceOnly       bool
  CancelReason     string
  Status           string
}

type CreateOrderParams struct {
  ClientId    string `json:"clientId"`
  Market      string `json:"market"`
  Side        string `json:"side"`
  Type        string `json:"type"`
  Price       string `json:"price"`
  Size        string `json:"size"`
  LimitFee    string `json:"limitFee"`
  TimeInForce string `json:"timeInForce"`
  PostOnly    bool   `json:"postOnly"`
  Expiration  string `json:"expiration"`
  Signature   string `json:"signature"`
}

func (r *OrdersRepository) Find(id string) (*models.Order, error) {
  var entity *models.Order
  result := r.Db.Take(&entity, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *OrdersRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Order{})
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["position_side"]; ok {
    query.Where("position_side", conditions["position_side"].(string))
  }
  if _, ok := conditions["status"]; ok {
    query.Where("status", conditions["status"].(string))
  }
  query.Count(&total)
  return total
}

func (r *OrdersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Order {
  var orders []*models.Order
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "order_id",
    "type",
    "position_side",
    "side",
    "price",
    "quantity",
    "open_time",
    "update_time",
    "reduce_only",
    "status",
  })
  if _, ok := conditions["symbol"]; ok {
    query.Where("symbol", conditions["symbol"].(string))
  }
  if _, ok := conditions["position_side"]; ok {
    query.Where("position_side", conditions["position_side"].(string))
  }
  if _, ok := conditions["status"]; ok {
    query.Where("status", conditions["status"].(string))
  }
  query.Order("open_time desc")
  query.Offset((current - 1) * pageSize).Limit(pageSize).Find(&orders)
  return orders
}

func (r *OrdersRepository) Lost(symbol string, side string, quantity float64, timestamp int64) string {
  var entity models.Order
  result := r.Db.Where("symbol=? AND side=? AND quantity=?", symbol, side, quantity).Order("updated_at desc").Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return ""
  }
  if entity.UpdatedAt.Unix() < timestamp {
    return ""
  }
  return entity.OrderId
}

func (r *OrdersRepository) Status(orderId string) string {
  var entity models.Order
  result := r.Db.Select("status").Where("order_id=?", orderId).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return ""
  }
  return entity.Status
}

func (r *OrdersRepository) Open(symbol string) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  q := url.Values{}
  q.Add("market", symbol)

  path := fmt.Sprintf("/v3/active-orders?%s", q.Encode())

  isoTimestamp := time.Unix(0, r.Timestamp()).UTC().Format("2006-01-02T15:04:05.000Z")
  payload := fmt.Sprintf("%sGET%s", isoTimestamp, path)

  secret, _ := base64.URLEncoding.DecodeString(os.Getenv("DYDX_TRADE_API_SECRET"))
  mac := hmac.New(sha256.New, secret)
  _, err = mac.Write([]byte(payload))
  if err != nil {
    return
  }
  signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

  url := fmt.Sprintf("%s%s", os.Getenv("DYDX_API_ENDPOINT"), path)
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("DYDX-SIGNATURE", signature)
  req.Header.Set("DYDX-API-KEY", os.Getenv("DYDX_TRADE_API_KEY"))
  req.Header.Set("DYDX-PASSPHRASE", os.Getenv("DYDX_TRADE_API_PASSPHRASE"))
  req.Header.Set("DYDX-TIMESTAMP", isoTimestamp)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var result map[string]interface{}
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["orders"]; !ok {
    err = errors.New("invalid response")
    return
  }

  for _, order := range result["orders"].([]interface{}) {
    data := order.(map[string]interface{})
    orderId := data["id"].(string)
    var entity models.Order
    result := r.Db.Where("order_id", orderId).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      r.Flush(orderId)
    }
  }
  return nil
}

func (r *OrdersRepository) Sync(symbol string, startTime int64, limit int) error {
  return nil
}

func (r *OrdersRepository) Create(
  symbol string,
  side string,
  price float64,
  quantity float64,
  positionSide string,
) (orderId string, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  path := "/v3/orders"
  expiration := time.Now().Add(20 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")

  params := &CreateOrderParams{
    ClientId:    helpers.RandomClientId(),
    Market:      symbol,
    Side:        side,
    Type:        "LIMIT",
    Price:       fmt.Sprintf("%v", price),
    Size:        fmt.Sprintf("%v", quantity),
    TimeInForce: "GTT",
    LimitFee:    "0.01",
  }
  params.Expiration = expiration

  params.Signature, err = starkex.OrderSign(os.Getenv("DYDX_STARK_PRIVATE_KEY"), starkex.OrderSignParam{
    NetworkId:  common.GetEnvInt("DYDX_NETWORK_ID"),
    PositionId: common.GetEnvInt64("DYDX_POSITION_ID"),
    ClientId:   params.ClientId,
    Market:     params.Market,
    Side:       params.Side,
    HumanPrice: params.Price,
    HumanSize:  params.Size,
    LimitFee:   params.LimitFee,
    Expiration: params.Expiration,
  })

  body, err := json.Marshal(params)
  if err != nil {
    return
  }

  isoTimestamp := time.Unix(0, r.Timestamp()).UTC().Format("2006-01-02T15:04:05.000Z")
  payload := fmt.Sprintf("%sPOST%s%s", isoTimestamp, path, string(body))
  log.Println("payload", payload)

  secret, _ := base64.URLEncoding.DecodeString(os.Getenv("DYDX_TRADE_API_SECRET"))
  mac := hmac.New(sha256.New, secret)
  _, err = mac.Write([]byte(payload))
  if err != nil {
    return
  }
  signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

  url := fmt.Sprintf("%s%s", os.Getenv("DYDX_API_ENDPOINT"), path)
  req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("DYDX-SIGNATURE", signature)
  req.Header.Set("DYDX-API-KEY", os.Getenv("DYDX_TRADE_API_KEY"))
  req.Header.Set("DYDX-PASSPHRASE", os.Getenv("DYDX_TRADE_API_PASSPHRASE"))
  req.Header.Set("DYDX-TIMESTAMP", isoTimestamp)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode < 200 || resp.StatusCode > 300 {
    body, _ := io.ReadAll(resp.Body)
    log.Println("response", string(body))
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

  data := result["order"].(map[string]interface{})

  if _, ok := result["order"]; !ok {
    err = errors.New("invalid response")
    return
  }

  order := &OrderInfo{}
  order.Symbol = data["market"].(string)
  order.OrderId = data["id"].(string)
  order.Type = data["type"].(string)
  order.PositionSide = positionSide
  order.Side = data["side"].(string)
  order.Price, _ = strconv.ParseFloat(data["price"].(string), 64)
  if data["triggerPrice"] != nil {
    order.ActivatePrice, _ = strconv.ParseFloat(data["triggerPrice"].(string), 64)
  }
  order.Quantity, _ = strconv.ParseFloat(data["size"].(string), 64)
  remainQuantity, _ := strconv.ParseFloat(data["remainingSize"].(string), 64)
  order.ExecutedQuantity = order.Quantity - remainQuantity
  openTime, _ := time.Parse("2006-01-02T15:04:05.000Z", data["createdAt"].(string))
  order.OpenTime = openTime.UnixMilli()
  if data["unfillableAt"] != nil {
    updateTime, _ := time.Parse("2006-01-02T15:04:05.000Z", data["unfillableAt"].(string))
    order.UpdateTime = updateTime.UnixMilli()
  }
  order.ReduceOnly = data["reduceOnly"].(bool)
  if data["cancelReason"] != nil {
    order.CancelReason = data["cancelReason"].(string)
  }
  order.Status = data["status"].(string)
  if order.Status == "PENDING" || order.Status == "OPEN" {
    if order.ExecutedQuantity > 0 {
      order.Status = "PARTIALLY_FILLED"
    } else {
      order.Status = "NEW"
    }
  }

  r.Save(order)

  return order.OrderId, nil
}

func (r *OrdersRepository) Take(
  symbol string,
  positionSide string,
  price float64,
) (orderId string, err error) {
  return
}

func (r *OrdersRepository) Stop(
  symbol string,
  positionSide string,
  price float64,
) (orderId string, err error) {
  return
}

func (r *OrdersRepository) Cancel(orderId string) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  path := fmt.Sprintf("/v3/orders/%s", orderId)

  isoTimestamp := time.Unix(0, r.Timestamp()).UTC().Format("2006-01-02T15:04:05.000Z")
  payload := fmt.Sprintf("%sDELETE%s", isoTimestamp, path)

  secret, _ := base64.URLEncoding.DecodeString(os.Getenv("DYDX_TRADE_API_SECRET"))
  mac := hmac.New(sha256.New, secret)
  _, err = mac.Write([]byte(payload))
  if err != nil {
    return
  }
  signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

  url := fmt.Sprintf("%s%s", os.Getenv("DYDX_API_ENDPOINT"), path)
  req, _ := http.NewRequest("DELETE", url, nil)
  req.Header.Set("DYDX-SIGNATURE", signature)
  req.Header.Set("DYDX-API-KEY", os.Getenv("DYDX_TRADE_API_KEY"))
  req.Header.Set("DYDX-PASSPHRASE", os.Getenv("DYDX_TRADE_API_PASSPHRASE"))
  req.Header.Set("DYDX-TIMESTAMP", isoTimestamp)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var result map[string]interface{}
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["cancelOrder"]; !ok {
    err = errors.New("invalid response")
    return
  }

  r.Flush(orderId)

  return
}

func (r *OrdersRepository) Flush(orderId string) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  path := fmt.Sprintf("/v3/orders/%s", orderId)

  isoTimestamp := time.Unix(0, r.Timestamp()).UTC().Format("2006-01-02T15:04:05.000Z")
  payload := fmt.Sprintf("%sGET%s", isoTimestamp, path)

  secret, _ := base64.URLEncoding.DecodeString(os.Getenv("DYDX_TRADE_API_SECRET"))
  mac := hmac.New(sha256.New, secret)
  _, err = mac.Write([]byte(payload))
  if err != nil {
    return
  }
  signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

  url := fmt.Sprintf("%s%s", os.Getenv("DYDX_API_ENDPOINT"), path)
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("DYDX-SIGNATURE", signature)
  req.Header.Set("DYDX-API-KEY", os.Getenv("DYDX_TRADE_API_KEY"))
  req.Header.Set("DYDX-PASSPHRASE", os.Getenv("DYDX_TRADE_API_PASSPHRASE"))
  req.Header.Set("DYDX-TIMESTAMP", isoTimestamp)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var result map[string]interface{}
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["order"]; !ok {
    err = errors.New("invalid response")
    return
  }

  data := result["order"].(map[string]interface{})

  order := &OrderInfo{}
  order.Symbol = data["market"].(string)
  order.OrderId = data["id"].(string)
  order.Type = data["type"].(string)
  order.Side = data["side"].(string)
  order.Price, _ = strconv.ParseFloat(data["price"].(string), 64)
  if data["triggerPrice"] != nil {
    order.ActivatePrice, _ = strconv.ParseFloat(data["triggerPrice"].(string), 64)
  }
  order.Quantity, _ = strconv.ParseFloat(data["size"].(string), 64)
  remainQuantity, _ := strconv.ParseFloat(data["remainingSize"].(string), 64)
  order.ExecutedQuantity = order.Quantity - remainQuantity
  openTime, _ := time.Parse("2006-01-02T15:04:05.000Z", data["createdAt"].(string))
  order.OpenTime = openTime.UnixMilli()
  if data["unfillableAt"] != nil {
    updateTime, _ := time.Parse("2006-01-02T15:04:05.000Z", data["unfillableAt"].(string))
    order.UpdateTime = updateTime.UnixMilli()
  }
  order.ReduceOnly = data["reduceOnly"].(bool)
  if data["cancelReason"] != nil {
    order.CancelReason = data["cancelReason"].(string)
  }
  order.Status = data["status"].(string)
  if order.Status == "PENDING" || order.Status == "OPEN" {
    if order.ExecutedQuantity > 0 {
      order.Status = "PARTIALLY_FILLED"
    } else {
      order.Status = "NEW"
    }
  }

  var entity models.Position
  res := r.Db.Where("symbol=? AND status=1", order.Symbol).Take(&entity)
  if errors.Is(res.Error, gorm.ErrRecordNotFound) {
    if order.Side == "BUY" {
      entity.Side = 1
    } else if order.Side == "SELL" {
      entity.Side = 2
    }
  }

  if entity.Side == 1 {
    order.PositionSide = "LONG"
  } else if entity.Side == 2 {
    order.PositionSide = "SHORT"
  }

  r.Save(order)

  return
}

func (r *OrdersRepository) Save(order *OrderInfo) error {
  var entity models.Order
  result := r.Db.Where("order_id", order.OrderId).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = models.Order{
      ID:               xid.New().String(),
      Symbol:           order.Symbol,
      OrderId:          order.OrderId,
      Type:             order.Type,
      PositionSide:     fmt.Sprintf("%v", order.PositionSide),
      Side:             order.Side,
      Price:            order.Price,
      ActivatePrice:    order.ActivatePrice,
      Quantity:         order.Quantity,
      ExecutedQuantity: order.ExecutedQuantity,
      OpenTime:         order.OpenTime,
      UpdateTime:       order.UpdateTime,
      ReduceOnly:       order.ReduceOnly,
      CancelReason:     order.CancelReason,
      Status:           order.Status,
    }
    r.Db.Create(&entity)
  } else {
    entity.ActivatePrice = order.ActivatePrice
    entity.ExecutedQuantity = order.ExecutedQuantity
    entity.UpdateTime = order.UpdateTime
    entity.CancelReason = order.CancelReason
    entity.Status = order.Status
    r.Db.Model(&models.Order{ID: entity.ID}).Updates(entity)
  }
  return nil
}

func (r *OrdersRepository) Timestamp() int64 {
  timestamp := time.Now().UnixMilli()
  value, err := r.Rdb.HGet(r.Ctx, "dydx:server", "timediff").Result()
  if err != nil {
    return timestamp
  }
  timediff, _ := strconv.ParseInt(value, 10, 64)
  return timestamp - timediff
}

func (r *OrdersRepository) Test() error {
  //ethereumAddress := os.Getenv("DYDX_ETHEREUM_ADDRESS")
  //client := dydx.New(types.Options{
  //  Host:                      os.Getenv("DYDX_API_ENDPOINT"),
  //  StarkPublicKey:            os.Getenv("DYDX_STARK_PUBLIC_KEY"),
  //  StarkPrivateKey:           os.Getenv("DYDX_STARK_PRIVATE_KEY"),
  //  StarkPublicKeyYCoordinate: os.Getenv("DYDX_STARK_PUBLIC_KEY_Y_COORDINATE"),
  //  DefaultEthereumAddress:    ethereumAddress,
  //  ApiKeyCredentials: &types.ApiKeyCredentials{
  //    Key:        os.Getenv("DYDX_TRADE_API_KEY"),
  //    Secret:     os.Getenv("DYDX_TRADE_API_SECRET"),
  //    Passphrase: os.Getenv("DYDX_TRADE_API_PASSPHRASE"),
  //  },
  //})
  //response, err := client.Private.CreateOrder(&private.ApiOrder{
  //  ApiBaseOrder: private.ApiBaseOrder{Expiration: helpers.ExpireAfter(5 * time.Minute)},
  //  Market:       "DOGE-USD",
  //  Side:         "BUY",
  //  Type:         "LIMIT",
  //  Size:         "100",
  //  Price:        "0.7355",
  //  ClientId:     helpers.RandomClientId(),
  //  TimeInForce:  "GTT",
  //  PostOnly:     false,
  //  LimitFee:     "0.02",
  //}, common.GetEnvInt64("DYDX_POSITION_ID"))
  //if err != nil {
  //  log.Println("error", err.Error())
  //  return err
  //}
  //log.Println("response", response, err)
  //account, err := client.Private.GetOrders(nil)
  //if err != nil {
  //  log.Fatal(err)
  //}
  //log.Println(account)
  return nil
}
