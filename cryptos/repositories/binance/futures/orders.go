package futures

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
  service "github.com/adshao/go-binance/v2/futures"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
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

func (r *OrdersRepository) Lost(symbol string, positionSide string, side string, quantity float64, timestamp int64) int64 {
  var entity models.Order
  result := r.Db.Where("symbol=? AND position_side=? AND side=? AND quantity=?", symbol, positionSide, side, quantity).Order("update_time desc").Take(&entity)
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

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
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
    return
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/openOrders", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/openOrders", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
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

  var result []*service.Order
  json.NewDecoder(resp.Body).Decode(&result)
  for _, order := range result {
    r.Save(order)
  }

  return
}

func (r *OrdersRepository) Sync(symbol string, startTime int64, limit int) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  if startTime > 0 {
    params.Add("startTime", fmt.Sprintf("%v", startTime))
  }
  params.Add("limit", fmt.Sprintf("%v", limit))
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
    return
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/allOrders", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/allOrders", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
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

  var result []*service.Order
  json.NewDecoder(resp.Body).Decode(&result)
  for _, order := range result {
    r.Save(order)
  }

  return
}

func (r *OrdersRepository) Create(
  symbol string,
  positionSide string,
  side string,
  price float64,
  quantity float64,
) (orderId int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("positionSide", positionSide)
  params.Add("side", side)
  params.Add("type", "LIMIT")
  params.Add("price", strconv.FormatFloat(price, 'f', -1, 64))
  params.Add("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
  params.Add("timeInForce", "GTC")
  params.Add("newOrderRespType", "RESULT")
  params.Add("recvWindow", "60000")

  log.Println("params", params)

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  data := url.Values{}

  var apiKey, apiSecret string
  var isTestNet bool
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    isTestNet = true
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TRADE_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TRADE_API_SECRET")
  }

  if len(apiSecret) == 64 {
    mac := hmac.New(sha256.New, []byte(apiSecret))
    _, err = mac.Write([]byte(payload))
    if err != nil {
      return
    }
    signature := mac.Sum(nil)
    data.Add("signature", fmt.Sprintf("%x", signature))
  } else {
    block, _ := pem.Decode([]byte(apiSecret))
    if block == nil {
      err = errors.New("invalid raa secret key")
      return
    }
    privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
      return 0, err
    }
    hashed := sha256.Sum256([]byte(payload))
    signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
    data.Add("signature", base64.StdEncoding.EncodeToString(signature))
  }

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  }

  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", apiKey)
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
  orderId = response.OrderID

  r.Flush(symbol, orderId)

  return
}

func (r *OrdersRepository) Take(
  symbol string,
  positionSide string,
  price float64,
) (orderId int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  var side string
  if positionSide == "LONG" {
    side = "SELL"
  } else {
    side = "BUY"
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("positionSide", positionSide)
  params.Add("side", side)
  params.Add("type", "TAKE_PROFIT_MARKET")
  params.Add("stopPrice", strconv.FormatFloat(price, 'f', -1, 64))
  params.Add("closePosition", "true")
  params.Add("timeInForce", "GTC")
  params.Add("newOrderRespType", "RESULT")
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  data := url.Values{}

  var apiKey, apiSecret string
  var isTestNet bool
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    isTestNet = true
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TRADE_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TRADE_API_SECRET")
  }

  if len(apiSecret) == 64 {
    mac := hmac.New(sha256.New, []byte(apiSecret))
    _, err = mac.Write([]byte(payload))
    if err != nil {
      return
    }
    signature := mac.Sum(nil)
    data.Add("signature", fmt.Sprintf("%x", signature))
  } else {
    block, _ := pem.Decode([]byte(apiSecret))
    if block == nil {
      err = errors.New("invalid raa secret key")
      return
    }
    privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
      return 0, err
    }
    hashed := sha256.Sum256([]byte(payload))
    signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
    data.Add("signature", base64.StdEncoding.EncodeToString(signature))
  }

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  }

  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", apiKey)
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
  orderId = response.OrderID

  return
}

func (r *OrdersRepository) Stop(
  symbol string,
  positionSide string,
  price float64,
) (orderId int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  var side string
  if positionSide == "LONG" {
    side = "SELL"
  } else {
    side = "BUY"
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("positionSide", positionSide)
  params.Add("side", side)
  params.Add("type", "STOP_MARKET")
  params.Add("stopPrice", strconv.FormatFloat(price, 'f', -1, 64))
  params.Add("closePosition", "true")
  params.Add("timeInForce", "GTC")
  params.Add("newOrderRespType", "RESULT")
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  data := url.Values{}

  var apiKey, apiSecret string
  var isTestNet bool
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    isTestNet = true
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TRADE_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TRADE_API_SECRET")
  }

  if len(apiSecret) == 64 {
    mac := hmac.New(sha256.New, []byte(apiSecret))
    _, err = mac.Write([]byte(payload))
    if err != nil {
      return
    }
    signature := mac.Sum(nil)
    data.Add("signature", fmt.Sprintf("%x", signature))
  } else {
    block, _ := pem.Decode([]byte(apiSecret))
    if block == nil {
      err = errors.New("invalid raa secret key")
      return
    }
    privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
      return 0, err
    }
    hashed := sha256.Sum256([]byte(payload))
    signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
    data.Add("signature", base64.StdEncoding.EncodeToString(signature))
  }

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  }

  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", apiKey)
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
  orderId = response.OrderID

  return
}

func (r *OrdersRepository) Cancel(symbol string, orderId int64) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("orderId", fmt.Sprintf("%v", orderId))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  data := url.Values{}

  var apiKey, apiSecret string
  var isTestNet bool
  if common.GetEnvInt("BINANCE_FUTURES_TESTNET_ENABLE") == 1 {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TESTNET_API_SECRET")
    isTestNet = true
  } else {
    apiKey = common.GetEnvString("BINANCE_FUTURES_TRADE_API_KEY")
    apiSecret = common.GetEnvString("BINANCE_FUTURES_TRADE_API_SECRET")
  }

  if len(apiSecret) == 64 {
    mac := hmac.New(sha256.New, []byte(apiSecret))
    _, err := mac.Write([]byte(payload))
    if err != nil {
      return err
    }
    signature := mac.Sum(nil)
    data.Add("signature", fmt.Sprintf("%x", signature))
  } else {
    block, _ := pem.Decode([]byte(apiSecret))
    if block == nil {
      return errors.New("invalid raa secret key")
    }
    privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
      return err
    }
    hashed := sha256.Sum256([]byte(payload))
    signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
    data.Add("signature", base64.StdEncoding.EncodeToString(signature))
  }

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  }

  req, _ := http.NewRequest("DELETE", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", apiKey)
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

  var response *binance.CancelOrderResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return
  }

  r.Flush(symbol, orderId)

  return
}

func (r *OrdersRepository) Flush(symbol string, orderId int64) (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("orderId", fmt.Sprintf("%v", orderId))
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
    return
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  var url string
  if isTestNet {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_TESTNET_API_ENDPOINT"))
  } else {
    url = fmt.Sprintf("%s/fapi/v1/order", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
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

  var order *service.Order
  json.NewDecoder(resp.Body).Decode(&order)

  r.Save(order)

  return
}

func (r *OrdersRepository) Save(order *service.Order) error {
  symbol := order.Symbol
  orderId := order.OrderID

  price, _ := strconv.ParseFloat(order.Price, 64)
  avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)
  stopPrice, _ := strconv.ParseFloat(order.StopPrice, 64)
  quantity, _ := strconv.ParseFloat(order.OrigQuantity, 64)
  executedQuantity, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

  entity, err := r.Get(symbol, orderId)
  if errors.Is(err, gorm.ErrRecordNotFound) {
    entity = &models.Order{
      ID:               xid.New().String(),
      Symbol:           symbol,
      OrderId:          orderId,
      Type:             fmt.Sprintf("%v", order.Type),
      PositionSide:     fmt.Sprintf("%v", order.PositionSide),
      Side:             fmt.Sprintf("%v", order.Side),
      Price:            price,
      AvgPrice:         avgPrice,
      StopPrice:        stopPrice,
      Quantity:         quantity,
      ExecutedQuantity: executedQuantity,
      OpenTime:         order.Time,
      UpdateTime:       order.UpdateTime,
      Status:           fmt.Sprintf("%v", order.Status),
      Remark:           "",
    }
    r.Db.Create(&entity)
  } else {
    values := map[string]interface{}{}
    if entity.AvgPrice != avgPrice {
      values["avg_price"] = avgPrice
    }
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
