package margin

import (
  "bytes"
  "context"
  "crypto"
  "crypto/rand"
  "crypto/rsa"
  "crypto/sha256"
  "crypto/x509"
  "encoding/base64"
  "encoding/json"
  "encoding/pem"
  "errors"
  "fmt"
  "github.com/adshao/go-binance/v2/common"
  "net"
  "net/http"
  "net/url"
  "os"
  "strconv"
  "time"

  "gorm.io/gorm"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"

  models "taoniu.local/cryptos/models/binance/spot/margin"
)

type OrdersRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *OrdersRepository) Lost(symbol string, side string, price float64, timestamp int64) int64 {
  var entity models.Order
  result := r.Db.Where("symbol=? AND side=? AND price=?", symbol, side, price).Order("updated_at desc").Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return 0
  }
  if entity.UpdatedAt.Unix() < timestamp {
    return 0
  }
  return entity.OrderID
}

func (r *OrdersRepository) Status(symbol string, orderID int64) string {
  var entity models.Order
  result := r.Db.Select("status").Where("symbol=? AND order_id=?", symbol, orderID).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return ""
  }
  return entity.Status
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

func (r *OrdersRepository) Create(
  symbol string,
  side string,
  price float64,
  quantity float64,
  isIsolated bool,
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
  if isIsolated {
    params.Add("isIsolated", "TRUE")
  }
  params.Add("timeInForce", "GTC")
  params.Add("newOrderRespType", "RESULT")
  params.Add("recvWindow", "60000")

  value, err := r.Rdb.HGet(r.Ctx, "binance:server", "timediff").Result()
  if err != nil {
    return
  }
  timediff, _ := strconv.ParseInt(value, 10, 64)

  timestamp := time.Now().UnixNano()/int64(time.Millisecond) - timediff
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
    apiErr := new(common.APIError)
    err = json.NewDecoder(resp.Body).Decode(&apiErr)
    if err == nil {
      return 0, apiErr
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

  var response binance.CreateOrderResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return
  }

  r.Flush(symbol, response.OrderID, true)

  return response.OrderID, nil
}

func (r *OrdersRepository) Cancel(id string) error {
  var order models.Order
  result := r.Db.Where("id", id).First(&order)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }
  if order.Status != "NEW" {
    return errors.New("order status not valid")
  }

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
  params.Add("symbol", order.Symbol)
  params.Add("orderId", fmt.Sprintf("%v", order.OrderID))
  if order.IsIsolated {
    params.Add("isIsolated", "TRUE")
  }
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixNano() / int64(time.Millisecond)
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_SPOT_TRADE_API_SECRET")))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return err
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
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    apiErr := new(common.APIError)
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

  var response binance.CancelMarginOrderResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return err
  }

  order.Status = string(response.Status)
  r.Db.Model(&models.Order{ID: order.ID}).Updates(order)
  return nil
}

func (r *OrdersRepository) Flush(symbol string, orderID int64, isIsolated bool) error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  order, err := client.NewGetMarginOrderService().Symbol(symbol).OrderID(orderID).IsIsolated(isIsolated).Do(r.Ctx)
  if err != nil {
    return err
  }
  r.Save(order)

  var isolated int8
  if isIsolated {
    isolated = 1
  } else {
    isolated = 0
  }
  r.Rdb.SRem(
    r.Ctx,
    "binance:spot:margin:orders:flush",
    fmt.Sprintf("%s,%d,%d", symbol, orderID, isolated),
  )

  return nil
}

func (r *OrdersRepository) Sync(symbol string, isIsolated bool, limit int) error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  yestoday := time.Now().Unix() - 86400
  orders, err := client.NewListMarginOrdersService().Symbol(
    symbol,
  ).IsIsolated(
    isIsolated,
  ).StartTime(
    yestoday * 1000,
  ).Limit(
    limit,
  ).Do(r.Ctx)
  if err != nil {
    return err
  }
  for _, order := range orders {
    r.Save(order)
  }

  return nil
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
    r.Flush(order.Symbol, order.OrderID, order.IsIsolated)
  }
  return nil
}

func (r *OrdersRepository) Save(order *binance.Order) error {
  symbol := order.Symbol
  orderID := order.OrderID

  price, _ := strconv.ParseFloat(order.Price, 64)
  stopPrice, _ := strconv.ParseFloat(order.StopPrice, 64)
  quantity, _ := strconv.ParseFloat(order.OrigQuantity, 64)
  executedQuantity, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

  var entity models.Order
  result := r.Db.Where(
    "symbol=? AND order_id=?",
    symbol,
    orderID,
  ).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = models.Order{
      ID:               xid.New().String(),
      Symbol:           symbol,
      OrderID:          orderID,
      Type:             string(order.Type),
      Side:             string(order.Side),
      Price:            price,
      StopPrice:        stopPrice,
      Quantity:         quantity,
      ExecutedQuantity: executedQuantity,
      OpenTime:         order.Time,
      UpdateTime:       order.UpdateTime,
      IsIsolated:       order.IsIsolated,
      Status:           string(order.Status),
      Remark:           "",
    }
    r.Db.Create(&entity)
  } else {
    entity.ExecutedQuantity = executedQuantity
    entity.UpdateTime = order.UpdateTime
    entity.Status = string(order.Status)
    r.Db.Model(&models.Order{ID: entity.ID}).Updates(entity)
  }

  return nil
}
