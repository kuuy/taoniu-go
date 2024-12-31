package savings

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
  "os"
  "strconv"
  "time"

  "github.com/adshao/go-binance/v2"
  "github.com/adshao/go-binance/v2/common"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/savings"
)

type ProductsRepository struct {
  Db  *gorm.DB
  Ctx context.Context
}

func (r *ProductsRepository) Flush() error {
  client := binance.NewClient(
    os.Getenv("BINANCE_REST_API_KEY"),
    os.Getenv("BINANCE_REST_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  var current int64 = 1
  for {
    products, err := client.NewListSavingsFlexibleProductsService().Current(current).Size(100).Do(r.Ctx)
    if err != nil {
      return err
    }
    if len(products) == 0 {
      break
    }
    for _, product := range products {
      r.Save(product)
    }
    current += 1
  }

  return nil
}

func (r *ProductsRepository) Get(asset string) (models.FlexibleProduct, error) {
  var entity models.FlexibleProduct
  result := r.Db.Where("asset", asset).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return entity, result.Error
  }
  return entity, nil
}

func (r *ProductsRepository) Purchase(productId string, amount float64) (int64, error) {
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
  params.Add("productId", productId)
  params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMicro()
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

  url := "https://api.binance.com/sapi/v1/lending/daily/purchase"
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

  var response binance.PurchaseSavingsFlexibleProductResponse
  err = json.NewDecoder(resp.Body).Decode(&response)
  if err != nil {
    return 0, err
  }
  return int64(response.PurchaseId), nil
}

func (r *ProductsRepository) Save(product *binance.SavingsFlexibleProduct) error {
  asset := product.Asset
  productID := product.ProductId

  avgAnnualInterestRate, _ := strconv.ParseFloat(product.AvgAnnualInterestRate, 64)
  dailyInterestPerThousand, _ := strconv.ParseFloat(product.DailyInterestPerThousand, 64)
  minPurchaseAmount, _ := strconv.ParseFloat(product.MinPurchaseAmount, 64)
  purchasedAmount, _ := strconv.ParseFloat(product.PurchasedAmount, 64)
  upLimit, _ := strconv.ParseFloat(product.UpLimit, 64)
  upLimitPerUser, _ := strconv.ParseFloat(product.UpLimitPerUser, 64)

  var entity models.FlexibleProduct
  result := r.Db.Where("asset", asset).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = models.FlexibleProduct{
      ID:                       xid.New().String(),
      Asset:                    asset,
      ProductId:                productID,
      AvgAnnualInterestRate:    avgAnnualInterestRate,
      DailyInterestPerThousand: dailyInterestPerThousand,
      MinPurchaseAmount:        minPurchaseAmount,
      PurchasedAmount:          purchasedAmount,
      UpLimit:                  upLimit,
      UpLimitPerUser:           upLimitPerUser,
      CanPurchase:              product.CanPurchase,
      CanRedeem:                product.CanRedeem,
      Featured:                 product.Featured,
      Status:                   fmt.Sprint(product.Status),
    }
    r.Db.Create(&entity)
  } else {
    entity.ProductId = productID
    entity.AvgAnnualInterestRate = avgAnnualInterestRate
    entity.DailyInterestPerThousand = dailyInterestPerThousand
    entity.MinPurchaseAmount = minPurchaseAmount
    entity.PurchasedAmount = purchasedAmount
    entity.UpLimit = upLimit
    entity.UpLimitPerUser = upLimitPerUser
    entity.CanPurchase = product.CanPurchase
    entity.CanRedeem = product.CanRedeem
    entity.Featured = product.Featured
    entity.Status = fmt.Sprint(product.Status)
    r.Db.Model(&models.FlexibleProduct{ID: entity.ID}).Updates(entity)
  }

  return nil
}
