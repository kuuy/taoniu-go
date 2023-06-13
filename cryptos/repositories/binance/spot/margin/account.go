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
  "log"
  "net"
  "net/http"
  "net/url"
  "strconv"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/adshao/go-binance/v2/common"
  "github.com/go-redis/redis/v8"

  binanceConfig "taoniu.local/cryptos/config/binance"
  config "taoniu.local/cryptos/config/binance/spot"
)

type AccountRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

func (r *AccountRepository) Flush() error {
  client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
  account, err := client.NewGetMarginAccountService().Do(r.Ctx)
  if err != nil {
    return err
  }
  for _, coin := range account.UserAssets {
    log.Println("coin:", coin)
    free, _ := strconv.ParseFloat(coin.Free, 64)
    if free <= 0.0 {
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:margin:balance:%s", coin.Asset))
      continue
    }
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf("binance:spot:margin:balance:%s", coin.Asset),
      map[string]interface{}{
        "free":      coin.Free,
        "locked":    coin.Locked,
        "borrowed":  coin.Borrowed,
        "interrest": coin.Interest,
        "net_asset": coin.NetAsset,
      },
    )
  }

  return nil
}

func (r *AccountRepository) Loan(
  asset string,
  symbol string,
  amount float64,
  isIsolated bool,
) (int64, error) {
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
  params.Add("asset", asset)
  if symbol != "" {
    params.Add("symbol", symbol)
  }
  params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
  if isIsolated {
    params.Add("isIsolated", "TRUE")
  }
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixNano() / int64(time.Millisecond)
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(binanceConfig.FUND_SECRET_KEY))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return 0, err
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := "https://api.binance.com/sapi/v1/margin/loan"
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", binanceConfig.FUND_API_KEY)
  resp, err := httpClient.Do(req)
  if err != nil {
    return 0, err
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
    return 0, err
  }

  var response binance.TransactionResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return 0, err
  }

  return response.TranID, nil
}

func (r *AccountRepository) Repay(
  asset string,
  symbol string,
  amount float64,
  isIsolated bool,
) (int64, error) {
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
  params.Add("asset", asset)
  if symbol != "" {
    params.Add("symbol", symbol)
  }
  params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
  if isIsolated {
    params.Add("isIsolated", "TRUE")
  }
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixNano() / int64(time.Millisecond)
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(binanceConfig.FUND_SECRET_KEY))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return 0, err
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := "https://api.binance.com/sapi/v1/margin/repay"
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", binanceConfig.FUND_API_KEY)
  resp, err := httpClient.Do(req)
  if err != nil {
    return 0, err
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
    return 0, err
  }

  var response binance.TransactionResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return 0, err
  }

  return response.TranID, nil
}
