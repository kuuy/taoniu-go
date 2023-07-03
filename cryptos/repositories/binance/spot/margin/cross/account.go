package cross

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
    netAsset, _ := strconv.ParseFloat(coin.NetAsset, 64)
    if netAsset == 0.0 {
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:margin:cross:balance:%s", coin.Asset))
      continue
    }
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf("binance:spot:margin:cross:balance:%s", coin.Asset),
      map[string]interface{}{
        "free":      coin.Free,
        "locked":    coin.Locked,
        "borrowed":  coin.Borrowed,
        "interest":  coin.Interest,
        "net_asset": coin.NetAsset,
      },
    )
  }

  return nil
}

func (r *AccountRepository) Balance(symbol string) (map[string]float64, error) {
  fields := []string{
    "free",
    "locked",
    "borrowed",
    "interest",
    "net_asset",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:spot:margin:cross:balance:%s",
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

func (r *AccountRepository) Transfer(
  asset string,
  side int,
  quantity float64,
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
  params.Add("amount", strconv.FormatFloat(quantity, 'f', -1, 64))
  params.Add("type", strconv.Itoa(side))
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

  url := "https://api.binance.com/sapi/v1/margin/transfer"
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
