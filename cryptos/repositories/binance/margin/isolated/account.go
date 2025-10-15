package isolated

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
  "os"
  "strconv"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/adshao/go-binance/v2/common"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  binanceModels "taoniu.local/cryptos/models/binance/spot"
)

type AccountRepository struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  TradingsRepository *TradingsRepository
}

func (r *AccountRepository) Flush() error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  account, err := client.NewGetIsolatedMarginAccountService().Do(r.Ctx)
  if err != nil {
    return err
  }
  for _, coin := range account.Assets {
    baseTotalAsset, _ := strconv.ParseFloat(coin.BaseAsset.TotalAsset, 64)
    quoteTotalAsset, _ := strconv.ParseFloat(coin.QuoteAsset.TotalAsset, 64)
    if baseTotalAsset <= 0.0 && quoteTotalAsset <= 0.0 {
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:margin:isolated:balance:%s", coin.Symbol))
      continue
    }
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf("binance:spot:margin:isolated:balance:%s", coin.Symbol),
      map[string]interface{}{
        "margin_ratio":      coin.MarginRatio,
        "liquidate_price":   coin.LiquidatePrice,
        "base_free":         coin.BaseAsset.Free,
        "base_locked":       coin.BaseAsset.Locked,
        "base_borrowed":     coin.BaseAsset.Borrowed,
        "base_interest":     coin.BaseAsset.Interest,
        "base_net_asset":    coin.BaseAsset.NetAsset,
        "base_total_asset":  coin.BaseAsset.TotalAsset,
        "quote_free":        coin.QuoteAsset.Free,
        "quote_locked":      coin.QuoteAsset.Locked,
        "quote_borrowed":    coin.QuoteAsset.Borrowed,
        "quote_interest":    coin.QuoteAsset.Interest,
        "quote_net_asset":   coin.QuoteAsset.NetAsset,
        "quote_total_asset": coin.QuoteAsset.TotalAsset,
      },
    )
  }

  return nil
}

func (r *AccountRepository) Balance(symbol string) (float64, float64, error) {
  fields := []string{
    "quote_free",
    "base_free",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:spot:margin:isolated:balance:%s",
      symbol,
    ),
    fields...,
  ).Result()
  for i := 0; i < len(fields); i++ {
    if data[i] == nil {
      return 0, 0, errors.New(fmt.Sprintf("[%s] isolated margin balance not exists", symbol))
    }
  }
  balance, _ := strconv.ParseFloat(data[0].(string), 64)
  quantity, _ := strconv.ParseFloat(data[1].(string), 64)

  return balance, quantity, nil
}

func (r *AccountRepository) Collect() error {
  symbols := r.TradingsRepository.Scan()
  for _, symbol := range symbols {
    var entity *binanceModels.Symbol
    result := r.Db.Select([]string{"base_asset", "quote_asset"}).Where("symbol", symbol).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      continue
    }
    var quantity float64 = 0
    val, err := r.Rdb.HGet(
      r.Ctx,
      fmt.Sprintf("binance:spot:balance:%s", entity.BaseAsset),
      "free",
    ).Result()
    if err == nil {
      quantity, _ = strconv.ParseFloat(val, 64)
    }
    if quantity <= 0 {
      continue
    }
    transferId, err := r.Transfer(
      entity.BaseAsset,
      symbol,
      "SPOT",
      "ISOLATED_MARGIN",
      quantity,
    )
    if err != nil {
      return err
    }
    log.Println("transferId", transferId)
  }
  return nil
}

func (r *AccountRepository) Liquidate() error {
  symbols := r.TradingsRepository.Scan()
  for _, symbol := range symbols {
    var entity *binanceModels.Symbol
    result := r.Db.Select([]string{"quote_asset"}).Where("symbol", symbol).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      continue
    }
    var free float64 = 0
    var borrowed float64 = 0
    var interest float64 = 0
    data, err := r.Rdb.HMGet(
      r.Ctx,
      fmt.Sprintf("binance:spot:margin:isolated:balance:%s", symbol),
      "quote_free",
      "quote_borrowed",
      "quote_interest",
    ).Result()
    if data[0] == nil || data[1] == nil || data[2] == nil {
      continue
    }
    free, _ = strconv.ParseFloat(data[0].(string), 64)
    borrowed, _ = strconv.ParseFloat(data[1].(string), 64)
    interest, _ = strconv.ParseFloat(data[2].(string), 64)
    if borrowed <= 0 {
      continue
    }
    if free < interest {
      continue
    }
    transferId, err := r.Repay(
      entity.QuoteAsset,
      symbol,
      borrowed+interest,
      true,
    )
    if err != nil {
      return err
    }
    log.Println("transferId", transferId)
  }
  return nil
}

func (r *AccountRepository) Transfer(
  asset string,
  symbol string,
  from string,
  to string,
  quantity float64,
) (transferId int64, err error) {
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
  params.Add("symbol", symbol)
  params.Add("transFrom", from)
  params.Add("transTo", to)
  params.Add("amount", strconv.FormatFloat(quantity, 'f', -1, 64))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_FUND_API_SECRET")))
  privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
  if err != nil {
    return 0, err
  }
  hashed := sha256.Sum256([]byte(payload))
  signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])

  data := url.Values{}
  data.Add("signature", base64.StdEncoding.EncodeToString(signature))

  body := bytes.NewBufferString(fmt.Sprintf("%s&%s", payload, data.Encode()))

  url := "https://api.binance.com/sapi/v1/margin/isolated/transfer"
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_FUND_API_KEY"))
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    apiErr := new(common.APIError)
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

func (r *AccountRepository) Loan(
  asset string,
  symbol string,
  amount float64,
  isIsolated bool,
) (transactionId int64, err error) {
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
  params.Add("symbol", symbol)
  params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
  if isIsolated {
    params.Add("isIsolated", "TRUE")
  }
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

  url := "https://api.binance.com/sapi/v1/margin/loan"
  req, _ := http.NewRequest("POST", url, body)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_FUND_API_KEY"))
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    apiErr := new(common.APIError)
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
  transactionId = response.TranID

  return
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
    Timeout:   5 * time.Second,
  }

  params := url.Values{}
  params.Add("asset", asset)
  params.Add("symbol", symbol)
  params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
  if isIsolated {
    params.Add("isIsolated", "TRUE")
  }
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  payload := fmt.Sprintf("%s&timestamp=%v", params.Encode(), timestamp)

  block, _ := pem.Decode([]byte(os.Getenv("BINANCE_FUND_API_SECRET")))
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
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_FUND_API_KEY"))
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

  var response *binance.TransactionResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return 0, err
  }
  return response.TranID, nil
}
