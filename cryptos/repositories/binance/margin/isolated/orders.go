package isolated

import (
  "bytes"
  "context"
  "crypto"
  "crypto/hmac"
  "crypto/rand"
  "crypto/rsa"
  "crypto/sha256"
  "crypto/x509"
  "encoding/base64"
  "encoding/json"
  "encoding/pem"
  "errors"
  "fmt"
  "log"
  "net"
  "net/http"
  "net/url"
  "os"
  "strconv"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/margin/isolated"
)

type OrdersRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *OrdersRepository) Find(id string) (order *models.Order, err error) {
  result := r.Db.Take(&order, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = result.Error
    return
  }
  return
}

func (r *OrdersRepository) Get(symbol string, orderId int64) (order *models.Order, err error) {
  result := r.Db.Take(&order, "symbol=? AND order_id=?", symbol, orderId)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    err = result.Error
    return
  }
  return
}

func (r *OrdersRepository) Gets(conditions map[string]interface{}) []*models.Order {
  var positions []*models.Order
  query := r.Db.Select([]string{
    "symbol",
    "order_id",
  })
  if _, ok := conditions["status"]; ok {
    query.Where("status IN ?", conditions["status"].([]string))
  } else {
    query.Where("status IN ?", []string{"NEW", "PARTIALLY_FILLED"})
  }
  query.Find(&positions)
  return positions
}

func (r *OrdersRepository) Update(order *models.Order, column string, value interface{}) (err error) {
  r.Db.Model(&order).Update(column, value)
  return nil
}

func (r *OrdersRepository) Updates(order *models.Order, values map[string]interface{}) (err error) {
  err = r.Db.Model(&order).Updates(values).Error
  return
}

func (r *OrdersRepository) Count(conditions map[string]interface{}) int64 {
  var total int64
  query := r.Db.Model(&models.Order{})
  if _, ok := conditions["symbols"]; ok {
    query.Where("symbol IN ?", conditions["symbols"].([]string))
  }
  query.Where("status IN ?", []string{"NEW"})
  query.Count(&total)
  return total
}

func (r *OrdersRepository) Listings(conditions map[string]interface{}, current int, pageSize int) []*models.Order {
  offset := (current - 1) * pageSize

  var orders []*models.Order
  query := r.Db.Select([]string{
    "id",
    "symbol",
    "side",
    "price",
    "quantity",
    "status",
    "created_at",
    "updated_at",
  })
  if _, ok := conditions["symbols"]; ok {
    query.Where("symbol IN ?", conditions["symbols"].([]string))
  }
  query.Where("status IN ?", []string{"NEW"})
  query.Order("created_at desc")
  query.Offset(offset).Limit(pageSize).Find(&orders)
  return orders
}

func (r *OrdersRepository) Lost(symbol string, side string, quantity float64, timestamp int64) int64 {
  var entity models.Order
  result := r.Db.Where("symbol=? AND side=? AND quantity=?", symbol, side, quantity).Order("update_time desc").Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return 0
  }
  if entity.UpdateTime < timestamp {
    return 0
  }
  return entity.OrderId
}

func (r *OrdersRepository) Status(symbol string, orderId int64) string {
  if orderId == 0 {
    return ""
  }
  var entity models.Order
  result := r.Db.Select("status").Where("symbol=? AND order_id=?", symbol, orderId).Take(&entity)
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
    Timeout:   time.Duration(5) * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  params.Add("timestamp", fmt.Sprintf("%v", timestamp))

  mac := hmac.New(sha256.New, []byte(os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET")))
  _, err = mac.Write([]byte(params.Encode()))
  if err != nil {
    return
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  url := fmt.Sprintf("%s/sapi/v1/margin/openOrders", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  req.URL.RawQuery = params.Encode()
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"))
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

  var result []*binance.Order
  json.NewDecoder(resp.Body).Decode(&result)
  for _, order := range result {
    r.Save(order)
  }
  return
}

func (r *OrdersRepository) Create(
  symbol string,
  side string,
  price float64,
  quantity float64,
) (orderId int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(5) * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("side", side)
  params.Add("type", "LIMIT")
  params.Add("price", strconv.FormatFloat(price, 'f', -1, 64))
  params.Add("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
  params.Add("timeInForce", "GTC")
  params.Add("newOrderRespType", "RESULT")
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET")))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := fmt.Sprintf("%s/sapi/v1/margin/order", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"))
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    var apiErr *common.BinanceAPIError
    err = json.NewDecoder(resp.Body).Decode(&apiErr)
    if err == nil {
      err = apiErr
      return
    }
  }

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

  var response *binance.CreateOrderResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return
  }

  log.Println("order place response", response)

  r.Flush(symbol, response.OrderID)

  return response.OrderID, nil
}

func (r *OrdersRepository) Cancel(symbol string, orderId int64) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(5) * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("orderId", fmt.Sprintf("%v", orderId))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMicro()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_SPOT_TRADE_API_SECRET")))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := fmt.Sprintf("%s/sapi/v1/margin/order", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("DELETE", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_SPOT_TRADE_API_KEY"))
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    var apiErr *common.BinanceAPIError
    err = json.NewDecoder(resp.Body).Decode(&apiErr)
    if err == nil {
      return apiErr
    }
  }

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var response *binance.CancelMarginOrderResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return
  }

  r.Flush(symbol, orderId)

  return
}

func (r *OrdersRepository) Flush(symbol string, orderId int64) error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  order, err := client.NewGetMarginOrderService().Symbol(symbol).OrderID(orderId).IsIsolated(false).Do(r.Ctx)
  if err != nil {
    return err
  }
  r.Save(order)

  return nil
}

func (r *OrdersRepository) Sync(symbol string, startTime int64, limit int) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(5) * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("isIsolated", "TRUE")
  if startTime > 0 {
    params.Add("startTime", fmt.Sprintf("%v", startTime))
  }
  params.Add("limit", fmt.Sprintf("%v", limit))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  params.Add("timestamp", fmt.Sprintf("%v", timestamp))

  mac := hmac.New(sha256.New, []byte(os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET")))
  _, err = mac.Write([]byte(params.Encode()))
  if err != nil {
    return
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  url := fmt.Sprintf("%s/sapi/v1/margin/allOrders", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  req.URL.RawQuery = params.Encode()
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"))
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

  var result []*binance.Order
  json.NewDecoder(resp.Body).Decode(&result)
  for _, order := range result {
    r.Save(order)
  }

  return
}

func (r *OrdersRepository) Fix(time time.Time, limit int) error {
  var orders []*models.Order
  r.Db.Select([]string{
    "symbol",
    "order_id",
    "is_isolated",
  }).Where(
    "updated_at < ? AND status IN ?",
    time,
    []string{
      "NEW",
      "PARTIALLY_FILLED",
    },
  ).Order(
    "updated_at asc",
  ).Limit(
    limit,
  ).Find(&orders)
  for _, order := range orders {
    r.Flush(order.Symbol, order.OrderId)
  }
  return nil
}

func (r *OrdersRepository) Save(order *binance.Order) error {
  symbol := order.Symbol
  orderId := order.OrderID

  price, _ := strconv.ParseFloat(order.Price, 64)
  stopPrice, _ := strconv.ParseFloat(order.StopPrice, 64)
  quantity, _ := strconv.ParseFloat(order.OrigQuantity, 64)
  executedQuantity, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

  entity, err := r.Get(symbol, orderId)
  if errors.Is(err, gorm.ErrRecordNotFound) {
    entity = &models.Order{
      ID:               xid.New().String(),
      Symbol:           symbol,
      OrderId:          orderId,
      Type:             string(order.Type),
      Side:             string(order.Side),
      Price:            price,
      StopPrice:        stopPrice,
      Quantity:         quantity,
      ExecutedQuantity: executedQuantity,
      OpenTime:         order.Time,
      UpdateTime:       order.UpdateTime,
      Status:           string(order.Status),
      Remark:           "",
    }
    r.Db.Create(&entity)
  } else {
    values := map[string]interface{}{}
    if entity.ExecutedQuantity != executedQuantity {
      values["executed_quantity"] = executedQuantity
    }
    if entity.UpdateTime != order.UpdateTime {
      values["update_time"] = order.UpdateTime
    }
    if entity.Status != fmt.Sprintf("%v", order.Status) {
      values["status"] = fmt.Sprintf("%v", order.Status)
    }
    if len(values) > 0 {
      r.Updates(entity, values)
    }
  }

  return nil
}
