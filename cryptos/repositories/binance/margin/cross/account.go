package cross

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
  "io"
  "log"
  "net"
  "net/http"
  "net/url"
  "os"
  "slices"
  "strconv"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
)

type AccountRepository struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Nats *nats.Conn
}

type Balance struct {
  Asset    string `json:"asset"`
  Free     string `json:"free"`
  Locked   string `json:"locked"`
  Borrowed string `json:"borrowed"`
  Interest string `json:"interest"`
}

type AccountInfo struct {
  Balances []Balance `json:"userAssets"`
}

func (r *AccountRepository) Flush() (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   3 * time.Second,
  }

  params := url.Values{}
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

  url := fmt.Sprintf("%s/sapi/v1/margin/account", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
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

  var account *AccountInfo
  json.NewDecoder(resp.Body).Decode(&account)

  oldCurrencies, _ := r.Rdb.SMembers(r.Ctx, "binance:margin:cross:currencies").Result()
  var currencies []string
  for _, coin := range account.Balances {
    free, _ := strconv.ParseFloat(coin.Free, 64)
    locked, _ := strconv.ParseFloat(coin.Locked, 64)
    borrowed, _ := strconv.ParseFloat(coin.Borrowed, 64)
    interest, _ := strconv.ParseFloat(coin.Interest, 64)
    if free <= 0.0 && borrowed <= 0.0 {
      r.Rdb.SRem(r.Ctx, "binance:margin:cross:currencies", coin.Asset)
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:margin:cross:balance:%s", coin.Asset))
      continue
    }
    r.Rdb.SAdd(r.Ctx, "binance:margin:cross:currencies", coin.Asset)
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf("binance:margin:cross:balance:%s", coin.Asset),
      map[string]interface{}{
        "free":     free,
        "locked":   locked,
        "borrowed": borrowed,
        "interest": interest,
      },
    )
    currencies = append(currencies, coin.Asset)
  }

  for _, currency := range oldCurrencies {
    if !slices.Contains(currencies, currency) {
      r.Rdb.SRem(r.Ctx, "binance:margin:cross:currencies", currency)
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:margin:cross:balance:%s", currency))
    }
  }

  return
}

func (r *AccountRepository) Balance(symbol string) (map[string]float64, error) {
  fields := []string{
    "free",
    "locked",
    "borrowed",
    "interest",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:margin:cross:balance:%s",
      symbol,
    ),
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

func (r *AccountRepository) Transfer(asset string, side int, quantity float64) (transferId int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("asset", asset)
  params.Add("amount", strconv.FormatFloat(quantity, 'f', -1, 64))
  params.Add("type", strconv.Itoa(side))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_FUND_API_SECRET")))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := fmt.Sprintf("%s/sapi/v1/margin/transfer", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_FUND_API_KEY"))
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

  var response *binance.TransactionResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return
  }
  transferId = response.TranID

  return
}

func (r *AccountRepository) Borrow(asset string, amount float64) (transferId int64, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("asset", asset)
  params.Add("type", "BORROW")
  params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_FUND_API_SECRET")))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := fmt.Sprintf("%s/sapi/v1/margin/borrow-repay", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_FUND_API_KEY"))
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

  var response *binance.TransactionResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return
  }
  transferId = response.TranID

  return
}
