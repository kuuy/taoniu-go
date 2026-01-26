package swap

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "slices"
  "time"

  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "github.com/go-redis/redis/v8"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/raydium/swap"
)

type MintsRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *MintsRepository) Symbols() []string {
  var symbols []string
  r.Db.Model(models.Mint{}).Select("symbol").Where("status", "1").Find(&symbols)
  return symbols
}

func (r *MintsRepository) Get(symbol string) (entity *models.Mint, err error) {
  err = r.Db.Where("symbol", symbol).Take(&entity).Error
  return
}

func (r *MintsRepository) GetByAddress(address string) (entity *models.Mint, err error) {
  err = r.Db.Where("address", address).Take(&entity).Error
  return
}

func (r *MintsRepository) Create(
  name string,
  symbol string,
  address string,
  decimals int,
  tags []string,
  status int,
) (id string, err error) {
  id = xid.New().String()
  mint := &models.Mint{
    ID:       id,
    Symbol:   symbol,
    Name:     name,
    Address:  address,
    Decimals: decimals,
    Tags:     tags,
    Status:   status,
  }
  err = r.Db.Create(&mint).Error
  return
}

func (r *MintsRepository) Update(mint *models.Mint, column string, value interface{}) (err error) {
  return r.Db.Model(&mint).Update(column, value).Error
}

func (r *MintsRepository) Updates(mint *models.Mint, values map[string]interface{}) (err error) {
  return r.Db.Model(&mint).Updates(values).Error
}

func (r *MintsRepository) Flush() (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("RAYDIUM_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=30s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   30 * time.Second,
  }

  url := "https://api-v3.raydium.io/mint/list"
  fmt.Printf("fetching mints from: %s\n", url)
  req, err := http.NewRequestWithContext(r.Ctx, "GET", url, nil)
  if err != nil {
    return err
  }
  resp, err := httpClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode)
  }

  var response MintsListingsResponse
  if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
    return fmt.Errorf("decode error: %v", err)
  }

  if !response.Success {
    return fmt.Errorf("failed to fetch tokens: api returned success = false")
  }

  mintCount := len(response.Data.MintList)
  if mintCount == 0 {
    return fmt.Errorf("empty mints list")
  }

  fmt.Printf("received %d mints from api\n", mintCount)

  var symbols []string
  oldSymbols := r.Symbols()
  fmt.Printf("currently have %d active symbols in database\n", len(oldSymbols))

  var createdCount, updatedCount int

  for _, mintInfo := range response.Data.MintList {
    var status int
    if len(mintInfo.Tags) == 0 {
      status = 1
    } else {
      status = 4
    }
    var entity *models.Mint
    entity, err = r.Get(mintInfo.Symbol)
    if errors.Is(err, gorm.ErrRecordNotFound) {
      entity = &models.Mint{
        ID:       xid.New().String(),
        Name:     mintInfo.Name,
        Symbol:   mintInfo.Symbol,
        Address:  mintInfo.Address,
        Decimals: mintInfo.Decimals,
        Tags:     mintInfo.Tags,
        Status:   status,
      }
      r.Db.Create(&entity)
      createdCount++
    } else {
      values := map[string]interface{}{}
      if entity.Name != mintInfo.Name {
        values["name"] = mintInfo.Name
      }
      if entity.Decimals != mintInfo.Decimals {
        values["decimals"] = mintInfo.Decimals
      }
      if !slices.Equal(entity.Tags, mintInfo.Tags) {
        values["tags"] = datatypes.NewJSONSlice(mintInfo.Tags)
      }
      if entity.Status != status {
        values["status"] = status
      }
      if len(values) > 0 {
        r.Updates(entity, values)
        updatedCount++
      }
    }
    symbols = append(symbols, mintInfo.Symbol)
  }
  fmt.Printf("flush completed: %d created, %d updated\n", createdCount, updatedCount)

  var badSymbols []string
  for _, oldSymbol := range oldSymbols {
    if !slices.Contains(symbols, oldSymbol) {
      badSymbols = append(badSymbols, oldSymbol)
    }
  }

  if len(badSymbols) > 0 {
    fmt.Printf("deactivating %d symbols no longer present in api\n", len(badSymbols))
    err = r.Db.Model(&models.Mint{}).Where("symbol IN ?", badSymbols).Update("status", 4).Error
  }

  return err
}
